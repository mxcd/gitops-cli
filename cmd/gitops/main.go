package main

import (
	"os"

	"github.com/TwiN/go-color"
	"github.com/mxcd/gitops-cli/internal/finalizer"
	"github.com/mxcd/gitops-cli/internal/k8s"
	"github.com/mxcd/gitops-cli/internal/kubernetes"
	"github.com/mxcd/gitops-cli/internal/patch"
	"github.com/mxcd/gitops-cli/internal/state"
	"github.com/mxcd/gitops-cli/internal/templating"
	"github.com/mxcd/gitops-cli/internal/util"
	log "github.com/sirupsen/logrus"
	"github.com/urfave/cli/v2"
)

var version = "development"

func main() {
	cli.VersionFlag = &cli.BoolFlag{
		Name:    "version",
		Aliases: []string{},
		Usage:   "print only the version",
	}

	app := &cli.App{
		Name:    "gitpos",
		Version: version,
		Usage:   "GitOps CLI",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    "root-dir",
				Value:   "",
				Usage:   "root directory of the git repository",
				EnvVars: []string{"GITOPS_ROOT_DIR"},
			},
			&cli.StringFlag{
				Name:    "kubeconfig",
				Aliases: []string{"k"},
				Value:   "",
				Usage:   "kubeconfig file to use for connecting to the Kubernetes cluster",
				EnvVars: []string{"KUBECONFIG", "GITOPS_KUBECONFIG"},
			},
			&cli.BoolFlag{
				Name:    "verbose",
				Aliases: []string{"v"},
				Usage:   "debug output",
				EnvVars: []string{"GITOPS_VERBOSE"},
			},
			&cli.BoolFlag{
				Name:    "very-verbose",
				Aliases: []string{"vv"},
				Usage:   "trace output",
				EnvVars: []string{"GITOPS_VERY_VERBOSE"},
			},
			&cli.BoolFlag{
				Name:    "cleartext",
				Usage:   "print secrets in cleartext to the console",
				EnvVars: []string{"GITOPS_CLEARTEXT"},
			},
			&cli.BoolFlag{
				Name:    "print",
				Usage:   "print secrets to the console",
				EnvVars: []string{"GITOPS_PRINT"},
			},
			&cli.BoolFlag{
				Name:    "show-unchanged",
				Usage:   "display unchanged secrets in the plan overview",
				EnvVars: []string{"GITOPS_SHOW_UNCHANGED"},
			},
			&cli.IntFlag{
				Name:    "parallelism",
				Aliases: []string{"p"},
				Value:   5,
				Usage:   "number of parallel operations for decrypting secrets and kubernetes operations",
				EnvVars: []string{"GITOPS_PARALLELISM"},
			},
		},
		Commands: []*cli.Command{
			{
				Name:    "secrets",
				Aliases: []string{"s"},
				Usage:   "GitOps managed secrets",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:    "dir",
						Aliases: []string{"d"},
						Value:   "",
						Usage:   "directory to limit secret discovery to",
						EnvVars: []string{"GITOPS_SECRETS_DIR"},
					},
				},
				Subcommands: []*cli.Command{
					{
						Name:    "apply",
						Aliases: []string{"a"},
						Usage:   "Push secrets into your infrastructure",
						Subcommands: []*cli.Command{
							{
								Name:    "kubernetes",
								Aliases: []string{"k8s"},
								Usage:   "Push secrets into a Kubernetes cluster",
								Flags: []cli.Flag{
									&cli.BoolFlag{
										Name:  "auto-approve",
										Usage: "apply the changes without prompting for approval",
									},
								},
								Action: func(c *cli.Context) error {
									initApplication(c)
									return kubernetes.ApplyKubernetes(c)
								},
							},
							{
								Name:  "vault",
								Usage: "Push secrets into vault",
								Action: func(c *cli.Context) error {
									log.Fatal("Not implemented yet")
									return nil
								},
							},
						},
					},
					{
						Name:    "plan",
						Aliases: []string{"p"},
						Usage:   "Plan the application of secrets into your infrastructure",
						Subcommands: []*cli.Command{
							{
								Name:    "kubernetes",
								Aliases: []string{"k8s"},
								Usage:   "Plan the application of secrets into a Kubernetes cluster",

								Action: func(c *cli.Context) error {
									initApplication(c)
									return kubernetes.PlanKubernetes(c)
								},
							},
						},
					},
					{
						Name:  "template",
						Usage: "Test the templating of secrets",
						Action: func(c *cli.Context) error {
							initApplication(c)
							return templating.TestTemplating(c)
						},
					},
				},
			},
			{
				Name:  "clusters",
				Usage: "Managing target clusters",
				Subcommands: []*cli.Command{
					{
						Name:  "list",
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
							return finalizer.ExitApplication(c, true)
						},
					},
					{
						Name:  "add",
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
								Name:       c.Args().Get(0),
								ConfigFile: kubeconfig,
							})
							if err != nil {
								return err
							}
							return finalizer.ExitApplication(c, true)
						},
					},
					{
						Name:  "remove",
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
							return finalizer.ExitApplication(c, true)
						},
					},
					{
						Name:  "test",
						Usage: "Test a target cluster connection",
						Action: func(c *cli.Context) error {
							initApplication(c)
							k8s.InitClusterClients(c)
							clusterClients := k8s.GetClients()
							for _, clusterClient := range clusterClients {
								clusterClient.PrettyPrint()
							}
							return finalizer.ExitApplication(c, false)
						},
					},
				},
			},
			{
				Name:  "patch",
				Usage: "Patch a single file in a GitOps cluster repository",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:    "repository-server",
						Value:   "",
						Usage:   "GitOps repo server to be used for the patch",
						EnvVars: []string{"GITOPS_REPOSITORY_SERVER"},
					},
					&cli.StringFlag{
						Name:    "repository-server-api-key",
						Value:   "",
						Usage:   "GitOps repo server to be used for the patch",
						EnvVars: []string{"GITOPS_REPOSITORY_SERVER_API_KEY"},
					},
					&cli.StringFlag{
						Name:    "repository",
						Value:   "",
						Usage:   "Repo to be used for the patch",
						EnvVars: []string{"GITOPS_REPOSITORY"},
					},
					&cli.StringFlag{
						Name:    "branch",
						Value:   "main",
						Usage:   "Branch to be used for the patch",
						EnvVars: []string{"GITOPS_REPOSITORY_BRANCH"},
					},
					&cli.StringFlag{
						Name:    "basicauth",
						Value:   "",
						Usage:   "BasicAuth for authenticating against the GitOps repository",
						EnvVars: []string{"GITOPS_REPOSITORY_BASICAUTH"},
					},
					&cli.StringFlag{
						Name:    "ssh-key",
						Value:   "",
						Usage:   "SSH key for authenticating against the GitOps repository",
						EnvVars: []string{"GITOPS_REPOSITORY_SSH_KEY"},
					},
					&cli.StringFlag{
						Name:    "ssh-key-file",
						Value:   "",
						Usage:   "SSH key file for authenticating against the GitOps repository",
						EnvVars: []string{"GITOPS_REPOSITORY_SSH_KEY_FILE"},
					},
					&cli.StringFlag{
						Name:    "ssh-key-passphrase",
						Value:   "",
						Usage:   "SSH key passphrase for authenticating against the GitOps repository",
						EnvVars: []string{"GITOPS_REPOSITORY_SSH_KEY_PASSPHRASE"},
					},
					&cli.StringFlag{
						Name:    "actor",
						Value:   "",
						Usage:   "Username of the actor to be used for the commit",
						EnvVars: []string{"GITOPS_ACTOR"},
					},
				},
				Action: func(c *cli.Context) error {
					initApplication(c)
					return patch.PatchCommand(c)
				},
			},
		},
	}

	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}

func initApplication(c *cli.Context) error {
	util.PrintLogo(c)
	util.SetLogLevel(c)
	util.SetCliContext(c)
	util.GetRootDir()
	err := state.LoadState(c)
	return err
}
