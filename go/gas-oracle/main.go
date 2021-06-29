package main

import (
	"fmt"
	"os"

	"github.com/ethereum-optimism/optimism/go/gas-oracle/flags"
	"github.com/ethereum-optimism/optimism/go/gas-oracle/oracle"
	"github.com/ethereum/go-ethereum/log"
	"github.com/urfave/cli"
)

func main() {
	app := cli.NewApp()
	app.Flags = flags.Flags

	app.Before = func(ctx *cli.Context) error {
		loglevel := ctx.GlobalUint64(flags.LogLevelFlag.Name)
		log.Root().SetHandler(log.LvlFilterHandler(log.Lvl(loglevel), log.StreamHandler(os.Stdout, log.TerminalFormat(true))))
		return nil
	}

	app.Action = func(ctx *cli.Context) error {
		if args := ctx.Args(); len(args) > 0 {
			return fmt.Errorf("invalid command: %q", args[0])
		}

		config := oracle.NewConfig(ctx)
		gpo, err := oracle.NewGasPriceOracle(config)
		if err != nil {
			return err
		}

		if err := gpo.Start(); err != nil {
			return err
		}
		gpo.Wait()

		return nil
	}

	err := app.Run(os.Args)
	if err != nil {
		log.Crit("application failed", "message", err)
	}
}
