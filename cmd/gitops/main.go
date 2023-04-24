package main

import (
	"os"
	"time"

	"github.com/TwiN/go-color"
	"github.com/mxcd/gitops-cli/internal/state"
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
			&cli.BoolFlag{
				Name:  "cleartext",
				Usage: "print secrets in cleartext to the console",
				EnvVars: []string{"GITOPS_CLEARTEXT"},
			},
			&cli.BoolFlag{
				Name:  "print",
				Usage: "print secrets to the console",
				EnvVars: []string{"GITOPS_PRINT"},
			},
		},
		Commands: []*cli.Command{
			{
				Name: "secrets",
				Aliases: []string{"s"},
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
			{
				Name: "clusters",
				Usage: "Managing target clusters",
				Subcommands: []*cli.Command{
					{
						Name: "list",
						Usage: "List all target clusters",
						Action: func(c *cli.Context) error {
							initApplication(c)
							clusters := state.GetState().GetClusters()
							if len(clusters) == 0 {
								println("No clusters configured")
								return nil
							}
							for _, cluster := range clusters {
								println(color.InBlue(cluster.Name), " => ", cluster.ConfigFile)
							}
							return exitApplication(c, true)
						},
					},
					{
						Name: "add",
						Usage: "Add a target cluster. <name> <configFile>",
						Action: func(c *cli.Context) error {
							initApplication(c)
							kubeconfig := ""
							if c.Args().Len() == 1 {
								println("Please enter location of kubeconfig file for new ", color.InBlue(c.Args().Get(0)), " cluster: ")
								kubeconfig = util.StringPrompt("kubeconfig file: ")
							} else if c.Args().Len() == 2 {
								kubeconfig = c.Args().Get(1)
							} else {
								log.Fatal("Usage: gitops clusters add <name> <configFile>")
							}
							err := state.GetState().AddCluster(&state.ClusterState{
								Name: c.Args().Get(0),
								ConfigFile: kubeconfig,
							})
							if err != nil {
								return err
							}
							return exitApplication(c, true)
						},
					},
					{
						Name: "remove",
						Usage: "Remove a target cluster",
						Action: func(c *cli.Context) error {
							initApplication(c)
							if c.Args().Len() != 1 {
								log.Fatal("Usage: gitops clusters remove <name>")
							}
							err := state.GetState().RemoveCluster(c.Args().Get(0))
							if err != nil {
								return err
							}
							return exitApplication(c, true)
						},
					},
					{
						Name: "test",
						Usage: "Test a target cluster connection",
						Action: func(c *cli.Context) error {
							initApplication(c)
							
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

