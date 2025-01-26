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

type GitConnectionOptions struct {
	Repository string
	Branch     string
	FilePath   string
	Auth       transport.AuthMethod
}

type GitConnection struct {
	Repository *git.Repository
	Options    *GitConnectionOptions
}

func NewGitConnection(options *GitConnectionOptions) (*GitConnection, error) {
	return &GitConnection{
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

func GetAuthFromSshKey(sshKey []byte, sshKeyPassphrase string) (transport.AuthMethod, error) {
	signer, err := ssh.ParsePrivateKey(sshKey)
	if err != nil {
		return nil, err
	}
	auth := &gitssh.PublicKeys{User: "git", Signer: signer}
	return auth, nil
}
