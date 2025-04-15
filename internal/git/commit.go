package git

import (
	"fmt"

	"github.com/ldez/go-git-cmd-wrapper/v2/add"
	"github.com/ldez/go-git-cmd-wrapper/v2/commit"
	"github.com/ldez/go-git-cmd-wrapper/v2/config"
	"github.com/ldez/go-git-cmd-wrapper/v2/git"
	"github.com/ldez/go-git-cmd-wrapper/v2/revparse"
	"github.com/ldez/go-git-cmd-wrapper/v2/types"
	"github.com/rs/zerolog/log"
)

func (c *Connection) Commit(files []string, message string) (string, error) {
	directory := c.Options.Directory
	if directory == "" {
		return "", fmt.Errorf("directory is not specified")
	}

	msg, err := git.Add(runGitIn(directory), add.PathSpec(files...))
	if err != nil {
		log.Error().Err(err).Str("output", msg).Msg("Failed to add files")
		return "", err
	}

	msg, err = git.Config(config.Entry("user.name", c.Options.Signature.Name), runGitIn(directory))
	if err != nil {
		log.Error().Err(err).Str("output", msg).Msg("Failed to set user.name")
		return "", err
	}

	msg, err = git.Config(config.Entry("user.email", c.Options.Signature.Email), runGitIn(directory))
	if err != nil {
		log.Error().Err(err).Str("output", msg).Msg("Failed to set user.email")
		return "", err
	}

	msg, err = git.Commit(runGitIn(directory), commit.Message(message))
	if err != nil {
		log.Error().Err(err).Str("output", msg).Msg("Failed to commit")
		return "", err
	}

	currentCommitId, err := git.RevParse(runGitIn(directory), revparse.Args("HEAD"))
	if err != nil {
		log.Error().Err(err).Str("output", currentCommitId).Msg("Failed to get current commit ID")
		return "", err
	}

	return currentCommitId, nil
}

func (c *Connection) HasChanges() (bool, error) {
	directory := c.Options.Directory
	if directory == "" {
		return false, fmt.Errorf("directory is not specified")
	}

	msg, err := git.Raw("update-index", runGitIn(directory), func(g *types.Cmd) {
		g.AddOptions("--refresh")
	})
	if err != nil {
		log.Error().Err(err).Str("output", msg).Msg("Failed to refresh index")
		return false, err
	}

	_, err = git.Raw("diff-files", runGitIn(directory), func(g *types.Cmd) {
		g.AddOptions("--quiet")
	})
	if err != nil {
		return true, nil
	}

	return false, nil
}
