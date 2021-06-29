package oracle

import (
	"crypto/ecdsa"
	"fmt"
	"math/big"
	"strings"

	"github.com/ethereum-optimism/optimism/go/gas-oracle/flags"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/log"
	"github.com/urfave/cli"
)

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
