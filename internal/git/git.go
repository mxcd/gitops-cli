package git

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"sync"

	"github.com/ldez/go-git-cmd-wrapper/v2/git"
	"github.com/ldez/go-git-cmd-wrapper/v2/types"
	"github.com/rs/zerolog/log"
	"golang.org/x/crypto/ssh"
)

type ConnectionOptions struct {
	Directory        string
	Repository       string
	Branch           string
	Authentication   *Authentication
	IgnoreSshHostKey bool
	SkipSslVerify    bool
	Signature        *Signature
}

type Authentication struct {
	BasicAuth   *BasicAuth
	SshKey      *SshKey
	SshUsername *string
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
	Options *ConnectionOptions
}

var lock = &sync.Mutex{}

func NewGitConnection(options *ConnectionOptions) (*Connection, error) {
	connection := &Connection{
		Options: options,
	}

	if options.Signature == nil {
		connection.Options.Signature = &Signature{
			Name:  "GitOps CLI CI User",
			Email: "gitops@example.com",
		}
	} else {
		connection.Options.Signature = options.Signature
	}

	return connection, nil
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

func runGitIn(path string) types.Option {
	return git.CmdExecutor(
		func(ctx context.Context, name string, debug bool, args ...string) (string, error) {
			cmd := exec.CommandContext(ctx, name, args...)
			cmd.Dir = path

			output, err := cmd.CombinedOutput()
			return strings.TrimSuffix(string(output), "\n"), err
		},
	)
}

func (c *Connection) provideSshAuthentication() (*os.File, error) {
	// use ssh authentication only if provided
	if c.Options.Authentication != nil && c.Options.Authentication.SshKey != nil {
		key := c.Options.Authentication.SshKey
		// TODO add support for passphrase and ssh-agent
		if key.Passphrase != nil {
			log.Panic().Msg("Passphrase support is not implemented")
		}

		privateKeyFile, err := os.CreateTemp("", "git-ssh-key-*")
		if err != nil {
			log.Error().Err(err).Msg("Error creating temporary file for SSH key")
			return nil, err
		}

		if _, err := privateKeyFile.Write([]byte(key.PrivateKey)); err != nil {
			log.Error().Err(err).Msg("Error writing SSH key to temporary file")
			os.Remove(privateKeyFile.Name())
			return nil, err
		}
		if err := privateKeyFile.Chmod(0600); err != nil {
			log.Error().Err(err).Msg("Error setting permissions on SSH key file")
			os.Remove(privateKeyFile.Name())
			return nil, err
		}

		strictHostKeyChecking := ""
		if c.Options.IgnoreSshHostKey {
			strictHostKeyChecking = "-o StrictHostKeyChecking=no"
		}

		sshCommand := fmt.Sprintf("ssh -i %s %s", privateKeyFile.Name(), strictHostKeyChecking)
		if err := os.Setenv("GIT_SSH_COMMAND", sshCommand); err != nil {
			log.Error().Err(err).Msg("Error setting GIT_SSH_COMMAND")
			os.Remove(privateKeyFile.Name())
			return nil, err
		}

		return privateKeyFile, nil
	}

	return nil, nil
}

func cleanSshAuthentication(privateKeyFile *os.File) {
	if privateKeyFile != nil {
		if err := os.Remove(privateKeyFile.Name()); err != nil {
			log.Error().Err(err).Msg("Error removing temporary SSH key file")
		}
	}
}
