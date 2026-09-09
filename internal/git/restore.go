package git

import (
	"fmt"

	"github.com/ldez/go-git-cmd-wrapper/v2/git"
	"github.com/ldez/go-git-cmd-wrapper/v2/types"
	"github.com/rs/zerolog/log"
)

// Restore discards working tree and index changes of the given repository
// relative file paths by running `git checkout HEAD -- <files>`. Restoring from
// HEAD instead of the index also discards changes that were already staged by a
// failed commit. It intentionally does not touch commits, so a local commit
// whose push failed is preserved and pushed by a subsequent pull/push cycle.
func (c *Connection) Restore(files []string) error {
	directory := c.Options.Directory
	if directory == "" {
		return fmt.Errorf("directory is not specified")
	}

	if len(files) == 0 {
		return nil
	}

	msg, err := git.Raw("checkout", runGitIn(directory), func(g *types.Cmd) {
		g.AddOptions("HEAD")
		g.AddOptions("--")
		for _, file := range files {
			g.AddOptions(file)
		}
	})
	if err != nil {
		log.Error().Err(err).Str("output", msg).Msg("Failed to restore files")
		return err
	}

	log.Debug().Msgf("Restored %d file(s) in working tree", len(files))

	return nil
}
