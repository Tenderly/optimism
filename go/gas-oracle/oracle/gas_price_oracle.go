package oracle

import (
	"context"
	"crypto/ecdsa"
	"errors"
	"fmt"
	"math/big"
	"strings"
	"time"

	"github.com/ethereum-optimism/optimism/go/gas-oracle/bindings"
	"github.com/ethereum-optimism/optimism/go/gas-oracle/flags"
	"github.com/ethereum-optimism/optimism/go/gas-oracle/gasprices"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/log"
	"github.com/urfave/cli"
)

var errInvalidSigningKey = errors.New("invalid signing key")
var errNoChainID = errors.New("no chain id provided")
var errNoPrivateKey = errors.New("no private key provided")

// GasPriceOracle manages a hot key that can update the L2 Gas Price
type GasPriceOracle struct {
	chainID   *big.Int
	ctx       context.Context
	stop      chan struct{}
	contract  *bindings.GasPriceOracle
	backend   DeployContractBackend
	gasPricer *gasprices.L2GasPricer
	config    *Config
}

// Start runs the GasPriceOracle
func (g *GasPriceOracle) Start() error {
	if g.config.chainID == nil {
		return errNoChainID
	}
	if g.config.privateKey == nil {
		return errNoPrivateKey
	}

	address := crypto.PubkeyToAddress(g.config.privateKey.PublicKey)
	log.Info("Starting Gas Price Oracle", "chain-id", g.chainID, "address", address.Hex())

	// Fetch the owner of the contract to check that the local key matches
	// the owner of the contract. If it doesn't match then nothing can be
	// accomplished.
	owner, err := g.contract.Owner(&bind.CallOpts{
		Context: g.ctx,
	})
	if err != nil {
		return err
	}

	if address != owner {
		log.Error("Signing key does not match contract owner", "signer", address.Hex(), "owner", owner.Hex())
		return errInvalidSigningKey
	}

	// TODO: Errors in this goroutine should write to an error channel
	// and be handled externally
	go g.Loop()

	return nil
}

func (g *GasPriceOracle) Stop() {
	close(g.stop)
}

func (g *GasPriceOracle) Wait() {
	<-g.stop
}

func (g *GasPriceOracle) Loop() {
	tip, err := g.backend.HeaderByNumber(g.ctx, nil)
	if err != nil {
		log.Crit("Cannot fetch tip", "message", err)
	}
	// Start at the tip
	epochStartBlockNumber := float64(tip.Number.Uint64())
	// getLatestBlockNumberFn is used by the GasPriceUpdater
	// to get the latest block number
	getLatestBlockNumberFn := wrapGetLatestBlockNumberFn(g.backend)
	// updateL2GasPriceFn is used by the GasPriceUpdater to
	// update the gas price
	updateL2GasPriceFn, err := wrapUpdateL2GasPriceFn(g.backend, g.config)
	if err != nil {
		log.Crit("error", "message", err)
	}

	gasPriceUpdater := gasprices.NewGasPriceUpdater(
		g.gasPricer,
		epochStartBlockNumber,
		g.config.averageBlockGasLimitPerEpoch,
		g.config.epochLengthSeconds,
		getLatestBlockNumberFn,
		updateL2GasPriceFn,
	)

	// Iterate once per epoch
	timer := time.NewTicker(time.Duration(g.config.epochLengthSeconds) * time.Second)
	for {
		select {
		case <-timer.C:
			log.Debug("polling", "time", time.Now())

			l2GasPrice, err := g.contract.GasPrice(&bind.CallOpts{
				Context: g.ctx,
			})
			if err != nil {
				log.Error("cannot get gas price", "message", err)
				continue
			}

			if err := gasPriceUpdater.UpdateGasPrice(); err != nil {
				log.Error("cannot update gas price", "message", err)
				continue
			}

			newGasPrice := gasPriceUpdater.GetGasPrice()
			log.Info("Updated gas price", "previous", l2GasPrice, "current", newGasPrice)
		case <-g.ctx.Done():
			g.Stop()
		}
	}
}

func NewGasPriceOracle(cfg *Config) (*GasPriceOracle, error) {
	client, err := ethclient.Dial(cfg.ethereumHttpUrl)
	if err != nil {
		return nil, err
	}

	// Ensure that we can actually connect
	t := time.NewTicker(5 * time.Second)
	for ; true; <-t.C {
		_, err := client.ChainID(context.Background())
		if err == nil {
			t.Stop()
			break
		}
	}

	address := cfg.gasPriceOracleAddress
	contract, err := bindings.NewGasPriceOracle(address, client)
	if err != nil {
		return nil, err
	}

	// Fetch the current gas price to use as the current price
	currentPrice, err := contract.GasPrice(&bind.CallOpts{
		Context: context.Background(),
	})
	if err != nil {
		return nil, err
	}
	log.Info("Starting gas price", "price", currentPrice)

	// Create a gas pricer for the gas price updater
	gasPricer := gasprices.NewGasPricer(
		float64(currentPrice.Uint64()),
		cfg.floorPrice,
		func() float64 {
			return cfg.targetGasPerSecond
		},
		cfg.maxPercentChangePerEpoch,
	)

	chainID := cfg.chainID
	if chainID == nil {
		log.Info("ChainID unset, fetching remote")
		chainID, err = client.ChainID(context.Background())
		if err != nil {
			return nil, err
		}
		cfg.chainID = chainID
	}

	if cfg.privateKey == nil {
		return nil, errNoPrivateKey
	}

	return &GasPriceOracle{
		chainID:   chainID,
		ctx:       context.Background(),
		stop:      make(chan struct{}),
		contract:  contract,
		gasPricer: gasPricer,

		config:  cfg,
		backend: client,
	}, nil
}

type Config struct {
	chainID                      *big.Int
	ethereumHttpUrl              string
	gasPriceOracleAddress        common.Address
	privateKey                   *ecdsa.PrivateKey
	gasPrice                     *big.Int
	floorPrice                   float64
	targetGasPerSecond           float64
	maxPercentChangePerEpoch     float64
	averageBlockGasLimitPerEpoch float64
	epochLengthSeconds           float64
	significantFactor            float64
}

func NewConfig(ctx *cli.Context) *Config {
	cfg := Config{
		gasPriceOracleAddress: common.HexToAddress("0x420000000000000000000000000000000000000F"),
		significantFactor:     0.05,
	}
	if ctx.GlobalIsSet(flags.EthereumHttpUrlFlag.Name) {
		cfg.ethereumHttpUrl = ctx.GlobalString(flags.EthereumHttpUrlFlag.Name)
	}
	if ctx.GlobalIsSet(flags.ChainIDFlag.Name) {
		chainID := ctx.GlobalUint64(flags.ChainIDFlag.Name)
		cfg.chainID = new(big.Int).SetUint64(chainID)
	}
	if ctx.GlobalIsSet(flags.GasPriceOracleAddressFlag.Name) {
		addr := ctx.GlobalString(flags.GasPriceOracleAddressFlag.Name)
		cfg.gasPriceOracleAddress = common.HexToAddress(addr)
	}
	if ctx.GlobalIsSet(flags.PrivateKeyFlag.Name) {
		hex := ctx.GlobalString(flags.PrivateKeyFlag.Name)
		if strings.HasPrefix(hex, "0x") {
			hex = hex[2:]
		}
		key, err := crypto.HexToECDSA(hex)
		if err != nil {
			log.Error(fmt.Sprintf("Option %q: %v", flags.PrivateKeyFlag.Name, err))
		}
		cfg.privateKey = key
	}
	if ctx.GlobalIsSet(flags.TransactionGasPriceFlag.Name) {
		gasPrice := ctx.GlobalUint64(flags.TransactionGasPriceFlag.Name)
		cfg.gasPrice = new(big.Int).SetUint64(gasPrice)
	}
	if ctx.GlobalIsSet(flags.FloorPriceFlag.Name) {
		cfg.floorPrice = ctx.GlobalFloat64(flags.FloorPriceFlag.Name)
	}
	if ctx.GlobalIsSet(flags.TargetGasPerSecondFlag.Name) {
		cfg.targetGasPerSecond = ctx.GlobalFloat64(flags.TargetGasPerSecondFlag.Name)
	} else {
		log.Crit("Missing config option", "option", flags.TargetGasPerSecondFlag.Name)
	}
	if ctx.GlobalIsSet(flags.MaxPercentChangePerEpochFlag.Name) {
		cfg.maxPercentChangePerEpoch = ctx.GlobalFloat64(flags.MaxPercentChangePerEpochFlag.Name)
	} else {
		log.Crit("Missing config option", "option", flags.MaxPercentChangePerEpochFlag.Name)
	}
	if ctx.GlobalIsSet(flags.AverageBlockGasLimitPerEpochFlag.Name) {
		cfg.averageBlockGasLimitPerEpoch = ctx.GlobalFloat64(flags.AverageBlockGasLimitPerEpochFlag.Name)
	} else {
		log.Crit("Missing config option", "option", flags.AverageBlockGasLimitPerEpochFlag.Name)
	}
	if ctx.GlobalIsSet(flags.EpochLengthSecondsFlag.Name) {
		cfg.epochLengthSeconds = ctx.GlobalFloat64(flags.EpochLengthSecondsFlag.Name)
	} else {
		log.Crit("Missing config option", "option", flags.EpochLengthSecondsFlag.Name)
	}
	if ctx.GlobalIsSet(flags.SignificantFactorFlag.Name) {
		cfg.significantFactor = ctx.GlobalFloat64(flags.SignificantFactorFlag.Name)
	}
	return &cfg
}
