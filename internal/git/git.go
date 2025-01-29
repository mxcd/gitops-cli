package git

import (
	"errors"
	"os"
	"strings"

	git2go "github.com/libgit2/git2go/v34"

	"golang.org/x/crypto/ssh"
)

type ConnectionOptions struct {
	Directory        string
	Repository       string
	Branch           string
	Authentication   *Authentication
	IgnoreSslHostKey bool
	PullRebase       bool
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

func (c *Connection) credentialsCallback(url string, usernameFromURL string, allowedTypes git2go.CredentialType) (*git2go.Cred, error) {

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
