package main

import (
	"fmt"
	"os"
	"os/signal"
	"runtime"

	me "github.com/kasplex-evm/kasplex-relayer"
	"github.com/kasplex-evm/kasplex-relayer/config"
	"github.com/kasplex-evm/kasplex-relayer/impl"
	"github.com/kasplex-evm/kasplex-relayer/log"
	"github.com/urfave/cli/v2"
)

func logVersion() {
	log.Infow("Starting application",
		"gitRevision", me.GitRev,
		"gitBranch", me.GitBranch,
		"goVersion", runtime.Version(),
		"built", me.BuildDate,
		"os/arch", fmt.Sprintf("%s/%s", runtime.GOOS, runtime.GOARCH),
	)
}

func Start(ctx *cli.Context) error {
	fmt.Println("start")
	c, err := config.Load(ctx.String(flagCfg))
	if err != nil {
		return err
	}
	log.Init(c.Log)
	logVersion()

	relayer, err := impl.NewRelayer(&c.Relayer)
	if err != nil {
		return err
	}
	relayer.Start()

	log.Info("started successfully.")

	ch := make(chan os.Signal, 1)
	signal.Notify(ch, os.Interrupt)
	<-ch

	relayer.Stop()

	log.Info("stopped gracefully.")

	return nil
}
