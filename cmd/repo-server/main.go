package main

import (
	"os"

	"github.com/go-git/go-git/v5/plumbing/transport"
	"github.com/rs/zerolog/log"

	"github.com/mxcd/gitops-cli/internal/git"
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

	err = gitConnection.Clone()
	if err != nil {
		log.Panic().Err(err).Msg("error cloning git repository")
	}

	server, err := server.NewServer(&server.RouterConfig{
		DevMode:    config.Get().Bool("DEV"),
		Port:       config.Get().Int("PORT"),
		ApiBaseUrl: config.Get().String("API_BASE_URL"),
		ApiKeys:    config.Get().StringArray("API_KEYS"),
	}, gitConnection)

	if err != nil {
		log.Panic().Err(err).Msg("error initializing server")
	}

	server.RegisterMiddlewares()
	server.RegisterRoutes()

	err = server.Run()
	if err != nil {
		log.Panic().Err(err).Msg("error running server")
	}
}

func getGitConnectionOptions() *git.ConnectionOptions {
	var auth transport.AuthMethod
	var err error
	if config.Get().String("GITOPS_REPOSITORY_BASICAUTH") != "" {
		auth, err = git.GetAuthFromBasicAuthString(config.Get().String("GITOPS_REPOSITORY_BASICAUTH"))
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

		auth, err = git.GetAuthFromSshKey(sshKey, passphrase, config.Get().Bool("GITOPS_REPOSITORY_NO_STRICT_HOST_KEY_CHECKING"))

		if err != nil {
			log.Panic().Err(err).Msg("error getting ssh key from string")
		}
	}

	return &git.ConnectionOptions{
		Branch:     config.Get().String("GITOPS_REPOSITORY_BRANCH"),
		Repository: config.Get().String("GITOPS_REPOSITORY"),
		Auth:       auth,
	}
}
