package patch

import (
	"errors"
	"io"
	"os"

	"github.com/go-git/go-git/v5/plumbing/transport"
	"github.com/mxcd/gitops-cli/internal/git"
	"github.com/mxcd/gitops-cli/internal/yaml"
	"github.com/urfave/cli/v2"

	log "github.com/sirupsen/logrus"
)

type GitPatcherOptions struct {
	Branch        string
	RepositoryUrl string

	BasicAuth        string
	SshPrivateKey    []byte
	SshKeyPassphrase string
}

type GitPatcher struct {
	Options       *GitPatcherOptions
	GitConnection *git.GitConnection
}

func GetGitPatcherOptionsFromCli(c *cli.Context) (*GitPatcherOptions, error) {
	options := &GitPatcherOptions{
		Branch:        c.String("branch"),
		RepositoryUrl: c.String("repo"),
	}

	if c.String("basicauth") != "" {
		options.BasicAuth = c.String("basicauth")
	}

	if c.String("ssh-key") != "" {
		options.SshPrivateKey = []byte(c.String("sshkey"))
	} else if c.String("ssh-key-file") != "" {
		file, err := os.ReadFile(c.String("ssh-key-file"))
		if err != nil {
			return nil, err
		}
		options.SshPrivateKey = file
	}
	if c.String("ssh-key-passphrase") != "" {
		options.SshKeyPassphrase = c.String("ssh-key-passphrase")
	}

	return options, nil
}

func NewGitPatcher(options *GitPatcherOptions) (*GitPatcher, error) {
	if options.BasicAuth == "" && len(options.SshPrivateKey) == 0 {
		return nil, errors.New("no authentication method specified")
	}

	return &GitPatcher{
		Options: options,
	}, nil
}

func (p *GitPatcher) Prepare() error {
	var authMethod transport.AuthMethod = nil

	if p.Options.BasicAuth != "" {
		auth, err := git.GetAuthFromBasicAuthString(p.Options.BasicAuth)
		if err != nil {
			return err
		}
		authMethod = auth
	} else if len(p.Options.SshPrivateKey) > 0 {
		auth, err := git.GetAuthFromSshKey([]byte(p.Options.SshPrivateKey), p.Options.SshKeyPassphrase)
		if err != nil {
			return err
		}
		authMethod = auth
	}

	gitConnection, err := git.NewGitConnection(&git.GitConnectionOptions{
		Repository: p.Options.RepositoryUrl,
		Branch:     p.Options.Branch,
		Auth:       authMethod,
	})

	if err != nil {
		return err
	}

	err = gitConnection.Clone()
	if err != nil {
		return err
	}

	p.GitConnection = gitConnection

	return nil
}

func (p *GitPatcher) Patch(patchTasks []PatchTask) error {

	worktree, err := p.GitConnection.Repository.Worktree()
	if err != nil {
		log.WithError(err).Error("Failed to get worktree")
		return err
	}

	for _, patchTask := range patchTasks {
		filePath := patchTask.FilePath

		fileStat, err := worktree.Filesystem.Stat(filePath)
		if err != nil {
			log.Error("Failed to stat file: ", err)
		}

		file, err := worktree.Filesystem.OpenFile(filePath, os.O_RDWR, fileStat.Mode())
		if err != nil {
			log.Error("Failed to open file: ", err)
			return err
		}
		defer file.Close()

		// Read the contents of the file
		fileContents, err := io.ReadAll(file)
		if err != nil {
			log.Error("Failed to read file: ", err)
			return err
		}

		log.Debug("original yaml file: ", string(fileContents))

		for _, patch := range patchTask.Patches {
			selector := patch.Selector
			value := patch.Value

			log.Debug("patching file with selector '", selector, "' and value '", value, "'")
			patchedData, err := yaml.PatchYaml(fileContents, selector, value)
			if err != nil {
				return err
			}
			log.Debug("patched yaml file:\n", string(patchedData))

			err = file.Truncate(0)
			if err != nil {
				log.WithError(err).Error("Failed to truncate file")
				return err
			}

			_, err = file.Seek(0, 0)
			if err != nil {
				log.WithError(err).Error("Failed to seek file")
				return err
			}

			_, err = file.Write(patchedData)
			if err != nil {
				log.Error("Failed to write file: ", err)
				return err
			}

			hasChanges, err := p.GitConnection.HasChanges()
			if err != nil {
				return err
			}

			if !hasChanges {
				log.Info("No changes detected, exiting")
				return nil
			}

			commitHash, err := p.GitConnection.Commit([]string{filePath}, "feat(gitops): patching "+filePath)
			if err != nil {
				return err
			}
			log.Info("Created patch commit: ", commitHash.String())
		}

	}

	err = p.GitConnection.Push(p.Options.Branch)

	return err
}
