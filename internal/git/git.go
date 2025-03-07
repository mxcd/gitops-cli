package git

import (
	"errors"
	"fmt"
	"os"
	"strings"

	git2go "github.com/libgit2/git2go/v34"
	"github.com/rs/zerolog/log"

	"golang.org/x/crypto/ssh"
)

type ConnectionOptions struct {
	Directory        string
	Repository       string
	Branch           string
	Authentication   *Authentication
	IgnoreSslHostKey bool
	Signature        *Signature
}

type Authentication struct {
	BasicAuth *BasicAuth
	SshKey    *SshKey
}

type BasicAuth struct {
	Username string
	Password string
}

type SshKey struct {
	PrivateKey []byte
	Passphrase *string
	Signer     *ssh.Signer
}

type Signature struct {
	Name  string
	Email string
}

type Connection struct {
	Repository *git2go.Repository
	Options    *ConnectionOptions
}

func NewGitConnection(options *ConnectionOptions) (*Connection, error) {
	connection := &Connection{
		Options: options,
	}

	if options.Directory != "" {
		stat, err := os.Stat(options.Directory)
		if err == nil {
			if stat.IsDir() {
				repository, err := git2go.OpenRepository(options.Directory)
				if err != nil {
					return nil, err
				}
				connection.Repository = repository
			}
		}
	}

	if options.Signature == nil {
		connection.Options.Signature = &Signature{
			Name:  "GitOps CI User",
			Email: "gitops@example.com",
		}
	} else {
		connection.Options.Signature = options.Signature
	}

	return connection, nil
}

func GetAuthFromUsernamePassword(username string, password string) (*Authentication, error) {
	return &Authentication{
		BasicAuth: &BasicAuth{
			Username: username,
			Password: password,
		},
	}, nil
}

func GetAuthFromBasicAuthString(basicAuth string) (*Authentication, error) {
	split := strings.Split(basicAuth, ":")
	if len(split) != 2 {
		return nil, errors.New("invalid basic auth string")
	}
	return &Authentication{
		BasicAuth: &BasicAuth{
			Username: split[0],
			Password: split[1],
		},
	}, nil
}

func GetAuthFromSshKey(sshKey []byte, sshKeyPassphrase *string) (*Authentication, error) {
	var signer ssh.Signer
	if sshKeyPassphrase != nil {
		_signer, err := ssh.ParsePrivateKeyWithPassphrase(sshKey, []byte(*sshKeyPassphrase))
		if err != nil {
			return nil, err
		}
		signer = _signer
	} else {
		_signer, err := ssh.ParsePrivateKey(sshKey)
		if err != nil {
			return nil, err
		}
		signer = _signer
	}

	return &Authentication{
		SshKey: &SshKey{
			PrivateKey: sshKey,
			Passphrase: sshKeyPassphrase,
			Signer:     &signer,
		},
	}, nil
}

func (c *Connection) credentialsCallback(url string, usernameFromURL string, allowedTypes git2go.CredentialType) (*git2go.Credential, error) {

	if c.Options.Authentication == nil {
		return git2go.NewCredentialDefault()
	}

	if c.Options.Authentication.BasicAuth != nil {
		return git2go.NewCredentialUserpassPlaintext(c.Options.Authentication.BasicAuth.Username, c.Options.Authentication.BasicAuth.Password)
	}

	if c.Options.Authentication.SshKey != nil {
		return git2go.NewCredentialSSHKeyFromSigner(usernameFromURL, *c.Options.Authentication.SshKey.Signer)
	}

	return git2go.NewCredentialDefault()
}

func (c *Connection) HasChanges() (bool, error) {
	if c.Repository == nil {
		return false, fmt.Errorf("repository is not initialized")
	}

	index, err := c.Repository.Index()
	if err != nil {
		return false, fmt.Errorf("error accessing repository index: %w", err)
	}

	if index.EntryCount() > 0 {
		log.Debug().Msg("Staged changes detected")
		return true, nil
	}

	statusList, err := c.Repository.StatusList(&git2go.StatusOptions{
		Show:  git2go.StatusShowIndexAndWorkdir,
		Flags: git2go.StatusOptIncludeUntracked,
	})
	if err != nil {
		return false, fmt.Errorf("error retrieving repository status: %w", err)
	}
	defer statusList.Free()

	statusCount, err := statusList.EntryCount()
	if err != nil {
		return false, fmt.Errorf("error getting status entry count: %w", err)
	}

	hasChanges := statusCount > 0
	log.Debug().Bool("hasChanges", hasChanges).Msg("Checked repository status")

	return hasChanges, nil
}
