package git

import (
	"errors"
	"strings"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/transport"
	"github.com/go-git/go-git/v5/plumbing/transport/http"
	gitssh "github.com/go-git/go-git/v5/plumbing/transport/ssh"
	"golang.org/x/crypto/ssh"
)

type ConnectionOptions struct {
	Repository string
	Branch     string
	Auth       transport.AuthMethod
}

type Connection struct {
	Repository *git.Repository
	Options    *ConnectionOptions
}

func NewGitConnection(options *ConnectionOptions) (*Connection, error) {
	return &Connection{
		Options: options,
	}, nil
}

func GetAuthFromUsernamePassword(username string, password string) (transport.AuthMethod, error) {
	return &http.BasicAuth{
		Username: username,
		Password: password,
	}, nil
}

func GetAuthFromBasicAuthString(basicAuth string) (transport.AuthMethod, error) {
	split := strings.Split(basicAuth, ":")
	if len(split) != 2 {
		return nil, errors.New("invalid basic auth string")
	}
	return &http.BasicAuth{
		Username: split[0],
		Password: split[1],
	}, nil
}

func GetAuthFromSshKey(sshKey []byte, sshKeyPassphrase *string, noStrictHostKeyChecking bool) (transport.AuthMethod, error) {

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

	auth := &gitssh.PublicKeys{User: "git", Signer: signer}

	if noStrictHostKeyChecking {
		auth.HostKeyCallback = ssh.InsecureIgnoreHostKey()
	}

	return auth, nil
}
