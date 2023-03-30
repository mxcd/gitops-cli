package main

import (
	"os"

	"github.com/mxcd/gitops-cli/internal/util"
	log "github.com/sirupsen/logrus"
	"github.com/urfave/cli/v2"
)

func initApplication(c *cli.Context) {
	setLogLevel(c)
	getRootDir(c)
	if c.String("root-dir") == "" {
		log.Fatal("No root directory specified")
	}
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

func getRootDir(c *cli.Context) {
	var rootDir string
	if c.String("root-dir") != "" {
		log.Trace("Using root-dir flag")
	} else {
		log.Trace("Using git repo root")
		var err error
		rootDir, err = util.GetGitRepoRoot()
		if err != nil {
			log.Fatal(err)
		}
		c.Set("root-dir", rootDir)
		log.Trace("Using root directory: '", rootDir, "'")

		_, err = os.Stat(c.String("root-dir"))
		if os.IsNotExist(err) {
			log.Fatal("Root directory does not exist")
		}
	}
}