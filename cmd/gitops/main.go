package main

import (
	"os"
	"time"

	"github.com/mxcd/gitops-cli/internal/util"
	"github.com/schollz/progressbar/v3"
	log "github.com/sirupsen/logrus"
	"github.com/urfave/cli/v2"
)

func main() {
	app := &cli.App{
		Name:  "gitpos",
		Usage: "GitOps CLI",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:  "root-dir",
				Value: "",
				Usage: "root directory of the git repository",
				EnvVars: []string{"GITOPS_ROOT_DIR"},
			},
		},
		Commands: []*cli.Command{
			{
				Name: "bar",
				Action: func(c *cli.Context) error {
					barTest()
					return nil
				},
			},
			{
				Name: "secrets",
				Usage: "GitOps managed secrets",
				Subcommands: []*cli.Command{
					{
						Name: "push",
						Usage: "Push secrets into your infrastructure",
						Subcommands: []*cli.Command{
							{
								Name: "kubernetes",
								Usage: "Push secrets into a Kubernetes cluster",
								Action: func(c *cli.Context) error {
									log.Fatal("Not implemented yet")
									return nil
								},
							},
							{
								Name: "vault",
								Usage: "Push secrets into vault",
								Action: func(c *cli.Context) error {
									log.Fatal("Not implemented yet")
									return nil
								},
							},
						},
					},
					{
						Name: "test",
						Usage: "Test the templating of secrets",
						Action: func(c *cli.Context) error {
							rootDir := getRootDir(c)
							secretFiles, err := util.GetSecretFiles(rootDir)
							if err != nil {
								log.Fatal(err)
							}
							for _, secretFile := range secretFiles {
								log.Info(secretFile)
								decryptedFile, err := util.DecryptFile(secretFile)
								if err != nil {
									log.Fatal(err)
								}
								log.Info(decryptedFile)
							}
							return nil
						},
					},
				},
			},
		},
	}

	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}

func GetGitRepoRoot() {
	panic("unimplemented")
}

func barTest() {
	bar := progressbar.NewOptions(100,
		progressbar.OptionEnableColorCodes(true),
		progressbar.OptionShowBytes(false),
		progressbar.OptionSetWidth(30),
		progressbar.OptionSetDescription("[cyan][1/3][reset] Writing moshable file..."),
		progressbar.OptionSetTheme(progressbar.Theme{
			Saucer:        "[green]=[reset]",
			SaucerHead:    "[green]>[reset]",
			SaucerPadding: " ",
			BarStart:      "[",
			BarEnd:        "]",
	}))
	for i := 0; i < 100; i++ {
		bar.Add(1)
		time.Sleep(40 * time.Millisecond)
	}
}