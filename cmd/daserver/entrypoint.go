package main

import (
	"fmt"

	"github.com/urfave/cli/v2"

	oplog "github.com/ethereum-optimism/optimism/op-service/log"
	"github.com/ethereum-optimism/optimism/op-service/opio"
	celestia "github.com/celestiaorg/op-plasma-celestia"
)

type Server interface {
	Start() error
	Stop() error
}

func StartDAServer(cliCtx *cli.Context) error {
	if err := CheckRequired(cliCtx); err != nil {
		return err
	}

	cfg := ReadCLIConfig(cliCtx)
	if err := cfg.Check(); err != nil {
		return err
	}

	logCfg := oplog.ReadCLIConfig(cliCtx)

	l := oplog.NewLogger(oplog.AppOut(cliCtx), logCfg)
	oplog.SetGlobalLogHandler(l.Handler())

	l.Info("Initializing Plasma DA server...")

	var server Server

	switch {
	case cfg.CelestiaEnabled():
		l.Info("Using celestia storage", "url", cfg.CelestiaConfig().URL)
		store := celestia.NewCelestiaStore(cfg.CelestiaConfig())
		server = celestia.NewCelestiaServer(cliCtx.String(ListenAddrFlagName), cliCtx.Int(PortFlagName), store, l)
	}

	if err := server.Start(); err != nil {
		return fmt.Errorf("failed to start the DA server")
	} else {
		l.Info("Started DA Server")
	}

	defer func() {
		if err := server.Stop(); err != nil {
			l.Error("failed to stop DA server", "err", err)
		}
	}()

	opio.BlockOnInterrupts()

	return nil
}
