package main

import (
	"os"
	"time"

	"github.com/rs/zerolog/log"

	"github.com/mxcd/gitops-cli/internal/git"
	"github.com/mxcd/gitops-cli/internal/patch"
	"github.com/mxcd/gitops-cli/internal/repo_server/server"
	"github.com/mxcd/gitops-cli/internal/repo_server/util"
	"github.com/mxcd/go-config/config"
)

func main() {

	if err := util.InitConfig(); err != nil {
		log.Panic().Err(err).Msg("error initializing config")
	}
	config.Print()

	if err := util.InitLogger(); err != nil {
		log.Panic().Err(err).Msg("error initializing logger")
	}

	gitConnection, err := git.NewGitConnection(getGitConnectionOptions())
	if err != nil {
		log.Panic().Err(err).Msg("error initializing git connection")
	}
	log.Info().Msg("Cloning git repository")
	startTime := time.Now()

	err = gitConnection.Clone()
	if err != nil {
		log.Panic().Err(err).Msg("error cloning git repository")
	}
	log.Info().Msgf("Cloned git repository in %s", time.Since(startTime))

	gitPatcher, err := patch.NewGitPatcher(&patch.GitPatcherOptions{
		GitConnection: gitConnection,
	})
	if err != nil {
		log.Panic().Err(err).Msg("error initializing git patcher")
	}

	err = gitPatcher.Prepare(&patch.PrepareOptions{Clone: false})
	if err != nil {
		log.Panic().Err(err).Msg("error preparing git patcher")
	}

	server, err := server.NewServer(&server.RouterOptions{
		DevMode:    config.Get().Bool("DEV"),
		Port:       config.Get().Int("PORT"),
		ApiBaseUrl: config.Get().String("API_BASE_URL"),
		ApiKeys:    config.Get().StringArray("API_KEYS"),
	}, gitPatcher)

	if err != nil {
		log.Panic().Err(err).Msg("error initializing server")
	}

	server.RegisterMiddlewares()
	server.RegisterRoutes()

	log.Info().Msgf("Starting server on port %d", config.Get().Int("PORT"))
	err = server.Run()
	if err != nil {
		log.Panic().Err(err).Msg("error running server")
	}
}

func getGitConnectionOptions() *git.ConnectionOptions {
	var authentication *git.Authentication
	var err error
	if config.Get().String("GITOPS_REPOSITORY_BASICAUTH") != "" {
		authentication, err = git.GetAuthFromBasicAuthString(config.Get().String("GITOPS_REPOSITORY_BASICAUTH"))
		if err != nil {
			log.Panic().Err(err).Msg("error getting basic auth from string")
		}
	} else if config.Get().String("GITOPS_REPOSITORY_SSH_KEY") != "" || config.Get().String("GITOPS_REPOSITORY_SSH_KEY_FILE") != "" {
		var sshKey []byte
		if config.Get().String("GITOPS_REPOSITORY_SSH_KEY") != "" {
			sshKey = []byte(config.Get().String("GITOPS_REPOSITORY_SSH_KEY"))
		} else {
			sshKey, err = os.ReadFile(config.Get().String("GITOPS_REPOSITORY_SSH_KEY_FILE"))
			if err != nil {
				log.Panic().Err(err).Msg("error reading ssh key file")
			}
		}

		var passphrase *string = nil
		if config.Get().String("GITOPS_REPOSITORY_SSH_KEY_PASSPHRASE") != "" {
			_passphrase := config.Get().String("GITOPS_REPOSITORY_SSH_KEY_PASSPHRASE")
			passphrase = &_passphrase
		}

		authentication, err = git.GetAuthFromSshKey(sshKey, passphrase)

		if err != nil {
			log.Panic().Err(err).Msg("error getting ssh key from string")
		}
	}

	return &git.ConnectionOptions{
		Branch:           config.Get().String("GITOPS_REPOSITORY_BRANCH"),
		Repository:       config.Get().String("GITOPS_REPOSITORY"),
		Authentication:   authentication,
		IgnoreSshHostKey: config.Get().Bool("GITOPS_REPOSITORY_IGNORE_SSL_HOSTKEY"),
	}
}
