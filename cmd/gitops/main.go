package main

import (
	"os"
	"time"

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
			&cli.StringFlag{
				Name:  "kubeconfig",
				Aliases: []string{"k"},
				Value: "",
				Usage: "kubeconfig file to use for connecting to the Kubernetes cluster",
				EnvVars: []string{"KUBECONFIG", "GITOPS_KUBECONFIG"},
			},
			&cli.BoolFlag{
				Name:  "verbose",
				Aliases: []string{"v"},
				Usage: "debug output",
				EnvVars: []string{"GITOPS_VERBOSE"},
			},
			&cli.BoolFlag{
				Name:  "very-verbose",
				Aliases: []string{"vv"},
				Usage: "trace output",
				EnvVars: []string{"GITOPS_VERY_VERBOSE"},
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
						Name: "apply",
						Aliases: []string{"a"},
						Usage: "Push secrets into your infrastructure",
						Subcommands: []*cli.Command{
							{
								Name: "kubernetes",
								Aliases: []string{"k8s"},
								Usage: "Push secrets into a Kubernetes cluster",
								Flags: []cli.Flag{
									&cli.BoolFlag{
										Name: "auto-approve",
										Usage: "apply the changes without prompting for approval",
									},
								},
								Action: func(c *cli.Context) error {
									initApplication(c)
									return applyKubernetes(c)
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
						Name: "plan",
						Aliases: []string{"p"},
						Usage: "Plan the application of secrets into your infrastructure",
						Subcommands: []*cli.Command{
							{
								Name: "kubernetes",
								Aliases: []string{"k8s"},
								Usage: "Plan the application of secrets into a Kubernetes cluster",

								Action: func(c *cli.Context) error {
									initApplication(c)
									return planKubernetes(c)
								},
							},
						},
					},
					{
						Name: "template",
						Usage: "Test the templating of secrets",
						Action: func(c *cli.Context) error {
							initApplication(c)
							return testTemplating(c)
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

func barTest() {
	bar := progressbar.NewOptions(100,
		progressbar.OptionEnableColorCodes(true),
		progressbar.OptionShowBytes(false),
		progressbar.OptionSetWidth(30),
		progressbar.OptionSetDescription("[cyan][1/3][reset] Writing moshable file..."),
		/*progressbar.OptionSetTheme(progressbar.Theme{
			Saucer:        "[green]=[reset]",
			SaucerHead:    "[green]>[reset]",
			SaucerPadding: " ",
			BarStart:      "[",
			BarEnd:        "]",
	})*/)
	for i := 0; i < 100; i++ {
		bar.Add(1)
		time.Sleep(40 * time.Millisecond)
	}
}

