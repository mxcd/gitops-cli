package patch

import (
	"fmt"
	"os"
	"path"
	"time"

	"github.com/mxcd/gitops-cli/internal/git"
	"github.com/mxcd/gitops-cli/internal/yaml"
	"github.com/urfave/cli/v2"

	"github.com/rs/zerolog/log"
)

type GitPatcherOptions struct {
	GitConnectionOptions *git.ConnectionOptions
	GitConnection        *git.Connection
}

type GitPatcher struct {
	Options       *GitPatcherOptions
	GitConnection *git.Connection
}

func GetGitConnectionOptionsFromCli(c *cli.Context) (*git.ConnectionOptions, error) {

	var authentication *git.Authentication = nil

	if c.String("basicauth") != "" {
		auth, err := git.GetAuthFromBasicAuthString(c.String("basicauth"))
		if err != nil {
			return nil, err
		}
		authentication = auth
	} else if c.String("ssh-key") != "" || c.String("ssh-key-file") != "" {
		var sshKey []byte
		if c.String("ssh-key") != "" {
			sshKey = []byte(c.String("ssh-key"))
		} else {
			file, err := os.ReadFile(c.String("ssh-key-file"))
			if err != nil {
				return nil, err
			}
			sshKey = file
		}

		var passphrase *string
		if c.String("ssh-key-passphrase") != "" {
			_passphrase := c.String("ssh-key-passphrase")
			passphrase = &_passphrase
		}

		auth, err := git.GetAuthFromSshKey(sshKey, passphrase)
		if err != nil {
			return nil, err
		}
		authentication = auth
	}

	options := &git.ConnectionOptions{
		Repository:     c.String("repository"),
		Branch:         c.String("branch"),
		Authentication: authentication,
	}

	return options, nil
}

func GetGitPatcherOptionsFromCli(c *cli.Context) (*GitPatcherOptions, error) {

	gitConnectionOptions, err := GetGitConnectionOptionsFromCli(c)
	if err != nil {
		return nil, err
	}

	options := &GitPatcherOptions{
		GitConnectionOptions: gitConnectionOptions,
	}

	return options, nil
}

func NewGitPatcher(options *GitPatcherOptions) (*GitPatcher, error) {
	return &GitPatcher{
		Options: options,
	}, nil
}

func (p *GitPatcher) Prepare(options *PrepareOptions) error {

	if p.Options.GitConnection != nil {
		p.GitConnection = p.Options.GitConnection
	} else {
		gitConnection, err := git.NewGitConnection(p.Options.GitConnectionOptions)
		if err != nil {
			return err
		}
		p.GitConnection = gitConnection
	}

	if options != nil && options.Clone {
		err := p.GitConnection.Clone()
		if err != nil {
			return err
		}
	}

	return nil
}

func (p *GitPatcher) Patch(patchTasks []PatchTask) error {

	err := p.GitConnection.Pull()
	if err != nil {
		return err
	}

	for _, patchTask := range patchTasks {
		relativeFilePath := patchTask.FilePath

		absoluteFilePath := path.Join(p.GitConnection.Options.Directory, relativeFilePath)

		fileStat, err := os.Stat(absoluteFilePath)
		if err != nil {
			log.Error().Err(err).Msg("Failed to stat file")
		}

		fileContents, err := os.ReadFile(absoluteFilePath)
		if err != nil {
			log.Error().Err(err).Msg("Failed to read file")
			return err
		}

		log.Debug().Msgf("original yaml file: %s", string(fileContents))

		for _, patch := range patchTask.Patches {
			selector := patch.Selector
			value := patch.Value

			log.Debug().Msgf("patching file with selector '%s' and value '%s'", selector, value)
			patchedData, err := yaml.PatchYaml(fileContents, selector, value)
			if err != nil {
				return err
			}
			log.Debug().Msgf("patched yaml file:\n%s", string(patchedData))

			fileContents = patchedData
		}

		err = os.WriteFile(absoluteFilePath, fileContents, fileStat.Mode())
		if err != nil {
			log.Error().Err(err).Msg("Failed to write file")
			return err
		}

		hasChanges, err := p.GitConnection.HasChanges()
		if err != nil {
			return err
		}

		if !hasChanges {
			log.Info().Msg("No changes detected, exiting")
			return nil
		}

		commitFooter := ""

		if patchTask.Actor != "" {
			commitFooter = fmt.Sprintf("\n\nTriggered by: %s", patchTask.Actor)
		}

		commitHash, err := p.GitConnection.Commit([]string{relativeFilePath}, fmt.Sprintf("feat(gitops): patching %s%s", relativeFilePath, commitFooter))
		if err != nil {
			return err
		}
		log.Info().Msgf("Created patch commit: %s", commitHash)
	}

	executePush := func() error {
		err := p.GitConnection.Pull()
		if err != nil {
			log.Error().Err(err).Msg("Error pulling prior to push")
			return err
		}
		return p.GitConnection.Push()
	}

	for i := 0; i < 3; i++ {
		err = executePush()
		if err == nil {
			break
		}
		time.Sleep(5 * time.Second)
	}

	return err
}
