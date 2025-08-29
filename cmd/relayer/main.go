package main

import (
	"fmt"
	"os"

	relayer "github.com/kasplex-evm/kasplex-relayer"
	"github.com/urfave/cli/v2"
)

const (
	appName = "relayer"
	flagCfg = "cfg"
)

func main() {
	app := cli.NewApp()
	app.Name = appName
	app.Version = "1.0"
	flags := []cli.Flag{
		&cli.StringFlag{
			Name:     flagCfg,
			Aliases:  []string{"c"},
			Usage:    "Configuration `FILE`",
			Value:    "./config.default.toml",
			Required: false,
		},
	}

	app.Commands = []*cli.Command{
		{
			Name:    "version",
			Aliases: []string{},
			Usage:   "Application version and build",
			Action:  versionCmd,
		},
		{
			Name:    "run",
			Aliases: []string{},
			Usage:   "Run the relayer",
			Action:  Start,
			Flags:   flags,
		},
	}

	err := app.Run(os.Args)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func versionCmd(*cli.Context) error {
	relayer.PrintVersion(os.Stdout)
	return nil
}
