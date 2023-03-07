package main

import (
	"github.com/mxcd/gitops-cli/internal/util"
	log "github.com/sirupsen/logrus"
	"github.com/urfave/cli/v2"
)

func getRootDir(c *cli.Context) string {
	var rootDir string
	if c.String("root-dir") != "" {
		log.Trace("Using root-dir flag")
		rootDir = c.String("root-dir")
	} else {
		log.Trace("Using git repo root")
		var err error
		rootDir, err = util.GetGitRepoRoot()
		if err != nil {
			log.Fatal(err)
		}
	}
	log.Trace("Using root directory: '", rootDir, "'")
	return rootDir
}