package git

import (
	"fmt"
	"time"

	"github.com/ldez/go-git-cmd-wrapper/v2/git"
	"github.com/ldez/go-git-cmd-wrapper/v2/pull"
	"github.com/rs/zerolog/log"
)

func (c *Connection) Pull() error {
	directory := c.Options.Directory
	if directory == "" {
		return fmt.Errorf("directory is not specified")
	}

	startTime := time.Now()

	lock.Lock()
	defer lock.Unlock()

	privateKeyFile, err := c.provideSshAuthentication()
	if err != nil {
		return err
	}
	defer cleanSshAuthentication(privateKeyFile)

	msg, err := git.Pull(
		runGitIn(directory),
		pull.Repository("origin"),
		pull.Refspec(c.Options.Branch),
		pull.Rebase("true"),
	)
	if err != nil {
		log.Error().Err(err).Str("output", msg).Msg("Failed to pull repository")
		return err
	}

	log.Debug().Msgf("Pulled from origin/%s in %d ms", c.Options.Branch, time.Since(startTime).Milliseconds())

	return nil
}
