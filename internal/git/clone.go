package git

import (
	"fmt"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/ldez/go-git-cmd-wrapper/v2/clone"
	"github.com/ldez/go-git-cmd-wrapper/v2/git"
	"github.com/rs/zerolog/log"
)

type CloneProtocol string

const (
	CloneProtocolSsh   CloneProtocol = "ssh"
	CloneProtocolHttps CloneProtocol = "https"
	CloneProtocolHttp  CloneProtocol = "http"
)

func (c *Connection) Clone() error {
	if c.Options.Directory == "" {
		directoryName, err := os.MkdirTemp(os.TempDir(), "gitops-repo-")
		if err != nil {
			log.Error().Err(err).Msg("Error creating temporary directory for cloning")
			return err
		}
		c.Options.Directory = directoryName
	} else {
		err := os.MkdirAll(c.Options.Directory, 0755)
		if err != nil {
			log.Error().Err(err).Msg("Error creating directory for cloning")
			return err
		}
	}

	repositoryUrl := c.Options.Repository
	cloneProtocol := CloneProtocolHttps

	if strings.HasPrefix(repositoryUrl, "ssh://") || strings.HasPrefix(repositoryUrl, "git@") {
		cloneProtocol = CloneProtocolSsh
	}

	if strings.HasPrefix(repositoryUrl, "http://") {
		cloneProtocol = CloneProtocolHttp
	}

	if strings.HasPrefix(repositoryUrl, "https://") {
		cloneProtocol = CloneProtocolHttps
	}

	if cloneProtocol == CloneProtocolHttp || cloneProtocol == CloneProtocolHttps {
		return c.CloneHttps(cloneProtocol)
	} else if cloneProtocol == CloneProtocolSsh {
		return c.CloneSsh(cloneProtocol)
	} else {
		return fmt.Errorf("unsupported clone protocol: %s", cloneProtocol)
	}

}

func (c *Connection) CloneSsh(cloneProtocol CloneProtocol) error {
	startTime := time.Now()

	repositoryUrl := c.Options.Repository
	repositoryBaseUrl := strings.TrimPrefix(repositoryUrl, "ssh://")
	repositoryBaseUrl = strings.TrimSuffix(repositoryBaseUrl, ".git")

	repositoryUrlSplit := strings.Split(repositoryBaseUrl, "@")
	sshUsername := "git"
	if c.Options.Authentication != nil && c.Options.Authentication.SshUsername != nil {
		sshUsername = *c.Options.Authentication.SshUsername
	}

	if len(repositoryUrlSplit) == 1 {
		repositoryUrl = fmt.Sprintf("ssh://%s@%s.git", sshUsername, repositoryUrlSplit[0])
	} else {
		repositoryUrl = fmt.Sprintf("ssh://%s@%s.git", repositoryUrlSplit[0], repositoryUrlSplit[1])
	}

	lock.Lock()
	defer lock.Unlock()

	log.Debug().Msgf("Using SSH URL: %s", repositoryUrl)

	privateKeyFile, err := c.provideSshAuthentication()
	if err != nil {
		return err
	}
	defer cleanSshAuthentication(privateKeyFile)

	msg, err := git.Clone(
		clone.Repository(repositoryUrl),
		clone.Directory(c.Options.Directory),
	)
	if err != nil {
		log.Error().Err(err).Str("output", msg).Msg("Failed to clone repository")
		return err
	}

	log.Debug().Msgf("Cloned repository %s on branch %s in %d ms", c.Options.Repository, c.Options.Branch, time.Since(startTime).Milliseconds())
	return nil
}

func (c *Connection) CloneHttps(cloneProtocol CloneProtocol) error {
	startTime := time.Now()

	repositoryUrl := c.Options.Repository
	repositoryBaseUrl := repositoryUrl

	if strings.HasPrefix(repositoryUrl, "http://") {
		repositoryBaseUrl = strings.TrimPrefix(repositoryUrl, "http://")
	}

	if strings.HasPrefix(repositoryUrl, "https://") {
		repositoryBaseUrl = strings.TrimPrefix(repositoryUrl, "https://")
	}

	if cloneProtocol == CloneProtocolHttp {
		repositoryUrl = fmt.Sprintf("http://%s", repositoryBaseUrl)
	} else if cloneProtocol == CloneProtocolHttps {
		repositoryUrl = fmt.Sprintf("https://%s", repositoryBaseUrl)
	}

	log.Debug().Msgf("Using http URL: %s", repositoryUrl)

	parsedUrl, err := url.Parse(repositoryUrl)
	if err != nil {
		log.Error().Err(err).Msg("Error parsing repository URL")
		return err
	}

	if c.Options.Authentication != nil && c.Options.Authentication.BasicAuth != nil {
		auth := c.Options.Authentication.BasicAuth
		parsedUrl.User = url.UserPassword(auth.Username, auth.Password)
	}

	repositoryUrl = parsedUrl.String()

	lock.Lock()
	defer lock.Unlock()

	if parsedUrl.Scheme == "https" && c.Options.SkipSslVerify {
		if err := os.Setenv("GIT_SSL_NO_VERIFY", "1"); err != nil {
			log.Error().Err(err).Msg("Error setting GIT_SSL_NO_VERIFY")
			return err
		}
	}

	msg, err := git.Clone(
		clone.Repository(repositoryUrl),
		clone.Directory(c.Options.Directory),
	)
	if err != nil {
		log.Error().Err(err).Str("output", msg).Msg("Failed to clone repository")
		return err
	}

	log.Debug().Msgf("Cloned repository %s on branch %s in %d ms", c.Options.Repository, c.Options.Branch, time.Since(startTime).Milliseconds())
	return nil
}
