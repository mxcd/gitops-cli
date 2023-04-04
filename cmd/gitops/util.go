package main

import (
	"github.com/mxcd/gitops-cli/internal/state"
	"github.com/mxcd/gitops-cli/internal/util"
	log "github.com/sirupsen/logrus"
	"github.com/urfave/cli/v2"
)

func initApplication(c *cli.Context) error  {
	printLogo(c)
	setLogLevel(c)
	util.SetCliContext(c)
	util.GetRootDir()
	err := state.LoadState(c)
	return err
}

func exitApplication(c *cli.Context, writeState bool) error {
	var err error
	if writeState {
		err = state.GetState().Save(c)
	}
	return err
}

func setLogLevel(c *cli.Context) {
	log.SetLevel(log.InfoLevel)
	
	if c.Bool("verbose") {
		log.SetLevel(log.DebugLevel)
	}

	if c.Bool("very-verbose") {
		log.SetLevel(log.TraceLevel)
	}
}

func printLogo(c *cli.Context) {
	if c.Bool("no-logo") {
		return
	}
	println("")
	println("__             _ __")
	println("\\ \\     ____ _(_) /_____  ____  _____")
	println(" \\ \\   / __ `/ / __/ __ \\/ __ \\/ ___/")
	println(" / /  / /_/ / / /_/ /_/ / /_/ (__  )")
	println("/_/   \\__, /_/\\__/\\____/ .___/____/")
	println("     /____/           /_/")
	println("")
}