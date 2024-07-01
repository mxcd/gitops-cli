package util

import (
	log "github.com/sirupsen/logrus"
	"github.com/urfave/cli/v2"
)

func SetLogLevel(c *cli.Context) {
	log.SetLevel(log.InfoLevel)

	if c.Bool("verbose") {
		log.SetLevel(log.DebugLevel)
	}

	if c.Bool("very-verbose") {
		log.SetLevel(log.TraceLevel)
	}
}

func PrintLogo(c *cli.Context) {
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
