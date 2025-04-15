package git

import (
	"fmt"
	"time"

	"github.com/ldez/go-git-cmd-wrapper/v2/git"
	"github.com/ldez/go-git-cmd-wrapper/v2/push"
	"github.com/rs/zerolog/log"
)

func (c *Connection) Push() error {
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

	msg, err := git.Push(
		runGitIn(directory),
		push.Remote("origin"),
		push.RefSpec(c.Options.Branch),
	)
	if err != nil {
		log.Error().Err(err).Str("output", msg).Msg("Failed to push to remote")
		return err
	}

	log.Debug().Msgf("Pushed to origin %s in %d ms", c.Options.Branch, time.Since(startTime).Milliseconds())

	return nil
}
