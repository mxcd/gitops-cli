package git

import (
	"fmt"
	"strings"

	"github.com/ldez/go-git-cmd-wrapper/v2/git"
	"github.com/ldez/go-git-cmd-wrapper/v2/revparse"
	"github.com/ldez/go-git-cmd-wrapper/v2/types"
	"github.com/rs/zerolog/log"
)

// RevParse resolves the given revision to its commit id.
func (c *Connection) RevParse(revision string) (string, error) {
	directory := c.Options.Directory
	if directory == "" {
		return "", fmt.Errorf("directory is not specified")
	}

	commitId, err := git.RevParse(runGitIn(directory), revparse.Args(revision))
	if err != nil {
		log.Error().Err(err).Str("output", commitId).Msgf("Failed to resolve revision %s", revision)
		return "", err
	}

	return strings.TrimSpace(commitId), nil
}

// CommitFiles returns the repository relative paths of all files touched by the
// given commit.
func (c *Connection) CommitFiles(commitId string) ([]string, error) {
	directory := c.Options.Directory
	if directory == "" {
		return nil, fmt.Errorf("directory is not specified")
	}

	msg, err := git.Raw("show", runGitIn(directory), func(g *types.Cmd) {
		g.AddOptions("--name-only")
		g.AddOptions("--format=")
		g.AddOptions(commitId)
	})
	if err != nil {
		log.Error().Err(err).Str("output", msg).Msgf("Failed to list files of commit %s", commitId)
		return nil, err
	}

	files := []string{}
	for _, line := range strings.Split(msg, "\n") {
		line = strings.TrimSpace(line)
		if line != "" {
			files = append(files, line)
		}
	}

	return files, nil
}
