package git

import (
	"fmt"

	git2go "github.com/libgit2/git2go/v34"

	"github.com/rs/zerolog/log"
)

func (c *Connection) Push() error {
	if c.Repository == nil {
		return fmt.Errorf("repository is not initialized")
	}

	remote, err := c.Repository.Remotes.Lookup("origin")
	if err != nil {
		log.Error().Err(err).Msg("Error looking up remote")
		return fmt.Errorf("error looking up remote: %w", err)
	}

	pushOptions := &git2go.PushOptions{
		RemoteCallbacks: git2go.RemoteCallbacks{
			CredentialsCallback: c.credentialsCallback,
		},
	}

	refspec := fmt.Sprintf("refs/heads/%s:refs/heads/%s", c.Options.Branch, c.Options.Branch)

	err = remote.Push([]string{refspec}, pushOptions)
	if err != nil {
		log.Error().Err(err).Msg("Error pushing to remote")
		return fmt.Errorf("error pushing to remote: %w", err)
	}

	log.Debug().Str("branch", c.Options.Branch).Msg("Push completed successfully")
	return nil
}
