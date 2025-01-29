package git

import (
	"os"
	"time"

	git2go "github.com/libgit2/git2go/v34"

	"github.com/rs/zerolog/log"
)

func (c *Connection) Clone() error {
	if c.Options.Directory == "" {
		directoryName, err := os.MkdirTemp(os.TempDir(), "gitops-repo-")
		if err != nil {
			log.Error().Err(err).Msg("Error creating temporary directory for cloning")
			return err
		}
		c.Options.Directory = directoryName
	} else {
		err := os.MkdirAll(c.Options.Directory, 0755)
		if err != nil {
			log.Error().Err(err).Msg("Error creating directory for cloning")
			return err
		}
	}

	startTime := time.Now()

	repo, err := git2go.Clone(c.Options.Repository, c.Options.Directory, &git2go.CloneOptions{
		CheckoutBranch:  c.Options.Branch,
		CheckoutOptions: git2go.CheckoutOptions{},
		FetchOptions:    c.getFetchOptions(FetchOptionUsageDefault),
	})
	if err != nil {
		log.Error().Err(err).Msgf("Error cloning repository %s on branch %s", c.Options.Repository, c.Options.Branch)
		return err
	}

	log.Debug().Msgf("Cloned repository %s on branch %s in %d", c.Options.Repository, c.Options.Branch, time.Since(startTime))
	c.Repository = repo

	err = repo.CheckoutHead(&git2go.CheckoutOptions{
		Strategy: git2go.CheckoutSafe,
	})
	if err != nil {
		log.Error().Err(err).Msg("Error checking out HEAD")
		return err
	}
	return nil
}
