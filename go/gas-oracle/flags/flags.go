package flags

import "github.com/urfave/cli"

var EthereumHttpUrlFlag = cli.StringFlag{
	Name:   "ethereum-http-url",
	Value:  "http://127.0.0.1:8545",
	Usage:  "Sequencer HTTP Endpoint",
	EnvVar: "GAS_PRICE_ORACLE_ETHEREUM_HTTP_URL",
}

var ChainIDFlag = cli.Uint64Flag{
	Name:   "chain-id",
	Usage:  "L2 Chain ID",
	EnvVar: "GAS_PRICE_ORACLE_CHAIN_ID",
}

var GasPriceOracleAddressFlag = cli.StringFlag{
	Name:   "gas-price-oracle-address",
	Usage:  "Address of OVM_GasPriceOracle",
	Value:  "0x420000000000000000000000000000000000000F",
	EnvVar: "GAS_PRICE_ORACLE_GAS_PRICE_ORACLE_ADDRESS",
}

var PrivateKeyFlag = cli.StringFlag{
	Name:   "private-key",
	Usage:  "Private Key corresponding to OVM_GasPriceOracle Owner",
	EnvVar: "GAS_PRICE_ORACLE_PRIVATE_KEY",
}

var TransactionGasPriceFlag = cli.Uint64Flag{
	Name:   "transaction-gas-price",
	Usage:  "Hardcoded tx.gasPrice",
	EnvVar: "GAS_PRICE_ORACLE_TRANSACTION_GAS_PRICE",
}

var LogLevelFlag = cli.IntFlag{
	Name:  "loglevel",
	Value: 3,
	Usage: "log level to emit to the screen",
}

var FloorPriceFlag = cli.Float64Flag{
	Name:  "floor-price",
	Value: 0,
	Usage: "gas price floor",
}

var TargetGasPerSecondFlag = cli.Float64Flag{
	Name:  "target-gas-per-second",
	Value: 0,
	Usage: "target gas per second",
}

var MaxPercentChangePerEpochFlag = cli.Float64Flag{
	Name:  "max-percent-change-per-epoch",
	Usage: "max percent change of gas price per second",
}

var AverageBlockGasLimitPerEpochFlag = cli.Float64Flag{
	Name:  "average-block-gas-limit-per-epoch",
	Value: 11_000_000,
	Usage: "average block gas limit per epoch",
}

var EpochLengthSecondsFlag = cli.Float64Flag{
	Name:  "epoch-length-seconds",
	Usage: "length of epochs in seconds",
}

var SignificantFactorFlag = cli.Float64Flag{
	Name:  "significant-factor",
	Value: 0.05,
	Usage: "only update when the gas price changes by more than this factor",
}

var WaitForReceiptFlag = cli.BoolFlag{
	Name:  "wait-for-receipt",
	Usage: "wait for receipts when sending transactions",
}

var Flags = []cli.Flag{
	EthereumHttpUrlFlag,
	ChainIDFlag,
	GasPriceOracleAddressFlag,
	PrivateKeyFlag,
	TransactionGasPriceFlag,
	LogLevelFlag,
	FloorPriceFlag,
	TargetGasPerSecondFlag,
	MaxPercentChangePerEpochFlag,
	AverageBlockGasLimitPerEpochFlag,
	EpochLengthSecondsFlag,
	SignificantFactorFlag,
	WaitForReceiptFlag,
}
