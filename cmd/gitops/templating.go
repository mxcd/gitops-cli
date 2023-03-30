package main

import (
	"github.com/mxcd/gitops-cli/internal/util"
	log "github.com/sirupsen/logrus"
	"github.com/urfave/cli/v2"
)

func testTemplating(c *cli.Context) error {
	secretFiles, err := util.GetSecretFiles(c.String("root-dir"))
	if err != nil {
		log.Fatal(err)
	}
	for _, secretFile := range secretFiles {
		log.Debug(secretFile)
		decryptedFile, err := util.DecryptFile(secretFile)
		if err != nil {
			log.Fatal(err)
		}
		log.Trace(string(decryptedFile))
	}
	return nil
}