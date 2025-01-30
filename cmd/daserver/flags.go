package main

import (
	"encoding/hex"
	"errors"
	"fmt"

	"github.com/urfave/cli/v2"

	celestia "github.com/celestiaorg/op-plasma-celestia"
	opservice "github.com/ethereum-optimism/optimism/op-service"
	oplog "github.com/ethereum-optimism/optimism/op-service/log"
)

const (
	ListenAddrFlagName        = "addr"
	PortFlagName              = "port"
	CelestiaServerFlagName    = "celestia.server"
	CelestiaAuthTokenFlagName = "celestia.auth-token"
	CelestiaNamespaceFlagName = "celestia.namespace"
)

const EnvVarPrefix = "OP_PLASMA_DA_SERVER"

func prefixEnvVars(name string) []string {
	return opservice.PrefixEnvVar(EnvVarPrefix, name)
}

var (
	ListenAddrFlag = &cli.StringFlag{
		Name:    ListenAddrFlagName,
		Usage:   "server listening address",
		Value:   "127.0.0.1",
		EnvVars: prefixEnvVars("ADDR"),
	}
	PortFlag = &cli.IntFlag{
		Name:    PortFlagName,
		Usage:   "server listening port",
		Value:   3100,
		EnvVars: prefixEnvVars("PORT"),
	}
	CelestiaServerFlag = &cli.StringFlag{
		Name:    CelestiaServerFlagName,
		Usage:   "celestia server endpoint",
		Value:   "http://localhost:26658",
		EnvVars: prefixEnvVars("CELESTIA_SERVER"),
	}
	CelestiaAuthTokenFlag = &cli.StringFlag{
		Name:    CelestiaAuthTokenFlagName,
		Usage:   "celestia auth token",
		Value:   "",
		EnvVars: prefixEnvVars("CELESTIA_AUTH_TOKEN"),
	}
	CelestiaNamespaceFlag = &cli.StringFlag{
		Name:    CelestiaNamespaceFlagName,
		Usage:   "celestia namespace",
		Value:   "",
		EnvVars: prefixEnvVars("CELESTIA_NAMESPACE"),
	}
)

var requiredFlags = []cli.Flag{
	ListenAddrFlag,
	PortFlag,
}

var optionalFlags = []cli.Flag{
	CelestiaServerFlag,
	CelestiaAuthTokenFlag,
	CelestiaNamespaceFlag,
}

func init() {
	optionalFlags = append(optionalFlags, oplog.CLIFlags(EnvVarPrefix)...)
	Flags = append(requiredFlags, optionalFlags...)
}

// Flags contains the list of configuration options available to the binary.
var Flags []cli.Flag

type CLIConfig struct {
	UseGenericComm    bool
	CelestiaEndpoint  string
	CelestiaAuthToken string
	CelestiaNamespace string
}

func ReadCLIConfig(ctx *cli.Context) CLIConfig {
	return CLIConfig{
		CelestiaEndpoint:  ctx.String(CelestiaServerFlagName),
		CelestiaAuthToken: ctx.String(CelestiaAuthTokenFlagName),
		CelestiaNamespace: ctx.String(CelestiaNamespaceFlagName),
	}
}

func (c CLIConfig) Check() error {
	if c.CelestiaEnabled() && (c.CelestiaEndpoint == "" || c.CelestiaAuthToken == "" || c.CelestiaNamespace == "") {
		return errors.New("all Celestia flags must be set")
	}
	if c.CelestiaEnabled() {
		if _, err := hex.DecodeString(c.CelestiaNamespace); err != nil {
			return err
		}
	}
	return nil
}

func (c CLIConfig) CelestiaConfig() celestia.CelestiaConfig {
	ns, _ := hex.DecodeString(c.CelestiaNamespace)
	return celestia.CelestiaConfig{
		URL:       c.CelestiaEndpoint,
		AuthToken: c.CelestiaAuthToken,
		Namespace: ns,
	}
}

func (c CLIConfig) CelestiaEnabled() bool {
	return !(c.CelestiaEndpoint == "" && c.CelestiaAuthToken == "" && c.CelestiaNamespace == "")
}

func CheckRequired(ctx *cli.Context) error {
	for _, f := range requiredFlags {
		if !ctx.IsSet(f.Names()[0]) {
			return fmt.Errorf("flag %s is required", f.Names()[0])
		}
	}
	return nil
}
