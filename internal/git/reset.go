package git

import (
	"fmt"

	"github.com/ldez/go-git-cmd-wrapper/v2/git"
	"github.com/ldez/go-git-cmd-wrapper/v2/types"
	"github.com/rs/zerolog/log"
)

// ResetToUpstream forces the clone back to the state of origin/<branch>. An
// in-progress rebase is aborted and local commits, staged and working tree
// changes are discarded. It is the single recovery path after a failed patch,
// so the long-running clone can always pull again.
func (c *Connection) ResetToUpstream() error {
	directory := c.Options.Directory
	if directory == "" {
		return fmt.Errorf("directory is not specified")
	}

	// fails when no rebase is in progress, which is the common case
	git.Raw("rebase", runGitIn(directory), func(g *types.Cmd) {
		g.AddOptions("--abort")
	})

	upstream := "origin/" + c.Options.Branch
	msg, err := git.Raw("reset", runGitIn(directory), func(g *types.Cmd) {
		g.AddOptions("--hard")
		g.AddOptions(upstream)
	})
	if err != nil {
		log.Error().Err(err).Str("output", msg).Msgf("Failed to reset clone to %s", upstream)
		return err
	}

	log.Debug().Msgf("Reset clone to %s", upstream)

	return nil
}

// RequireTracked fails if any of the given repository relative paths is not
// tracked by git.
func (c *Connection) RequireTracked(files []string) error {
	directory := c.Options.Directory
	if directory == "" {
		return fmt.Errorf("directory is not specified")
	}

	msg, err := git.Raw("ls-files", runGitIn(directory), func(g *types.Cmd) {
		g.AddOptions("--error-unmatch")
		g.AddOptions("--")
		for _, file := range files {
			g.AddOptions(file)
		}
	})
	if err != nil {
		return fmt.Errorf("file is not tracked by git: %s", msg)
	}

	return nil
}
