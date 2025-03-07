package git

import (
	"fmt"
	"time"

	git2go "github.com/libgit2/git2go/v34"

	"github.com/rs/zerolog/log"
)

func (c *Connection) Pull() error {
	if c.Repository == nil {
		return fmt.Errorf("repository is not initialized")
	}

	fetchOptions := c.getFetchOptions(FetchOptionUsageDefault)

	err := c.fetchRemote(fetchOptions)
	if err != nil {
		log.Error().Err(err).Msg("Error during fetch")
		return err
	}

	err = c.performPullAction()
	if err != nil {
		log.Error().Err(err).Msg("Error during pull action")
		return err
	}
	log.Debug().Msg("Pull operation completed successfully.")

	head, err := c.Repository.Head()
	if err != nil {
		log.Error().Err(err).Msg("Error getting HEAD reference")
		return err
	}

	headCommit, err := c.Repository.LookupCommit(head.Target())
	if err != nil {
		log.Error().Err(err).Msg("Error looking up HEAD commit")
		return err
	}

	err = c.Repository.ResetToCommit(headCommit, git2go.ResetHard, &git2go.CheckoutOptions{})
	if err != nil {
		log.Error().Err(err).Msg("Error checking out HEAD")
		return err
	}
	return nil
}

func (c *Connection) fetchRemote(fetchOptions git2go.FetchOptions) error {
	remote, err := c.Repository.Remotes.Lookup("origin")
	if err != nil {
		return fmt.Errorf("error looking up remote: %w", err)
	}

	err = remote.Fetch([]string{fmt.Sprintf("refs/heads/%s:refs/remotes/origin/%s", c.Options.Branch, c.Options.Branch)}, &fetchOptions, "")
	if err != nil {
		return fmt.Errorf("error fetching from remote: %w", err)
	}

	log.Debug().Msg("Fetch completed successfully.")
	return nil
}

func (c *Connection) performPullAction() error {
	localBranchRef := fmt.Sprintf("refs/heads/%s", c.Options.Branch)
	localBranch, err := c.Repository.References.Lookup(localBranchRef)
	if err != nil {
		return fmt.Errorf("error looking up local branch: %w", err)
	}

	remoteBranchRef := fmt.Sprintf("refs/remotes/origin/%s", c.Options.Branch)
	remoteBranch, err := c.Repository.References.Lookup(remoteBranchRef)
	if err != nil {
		return fmt.Errorf("error looking up remote branch: %w", err)
	}

	// If both branches are equal, nothing to do.
	if localBranch.Target().Equal(remoteBranch.Target()) {
		log.Debug().Msg("nothing to pull. already up to date")
		return nil
	}

	// Check if local branch is ahead of remote.
	isLocalAhead, err := c.Repository.DescendantOf(localBranch.Target(), remoteBranch.Target())
	if err != nil {
		return fmt.Errorf("error checking if local is ahead of remote: %w", err)
	}
	if isLocalAhead {
		log.Debug().Msg("local branch is ahead of remote; nothing to pull")
		return nil
	}

	// Check if remote is ahead of local (i.e. fast-forward is possible).
	isRemoteAhead, err := c.Repository.DescendantOf(remoteBranch.Target(), localBranch.Target())
	if err != nil {
		return fmt.Errorf("error checking if fast-forward is possible: %w", err)
	}
	log.Debug().Bool("isRemoteAhead", isRemoteAhead).Msg("Checking if fast-forward is possible")

	if isRemoteAhead {
		return c.fastForward(localBranch, remoteBranch)
	} else {
		return c.rebase(localBranch, remoteBranch)
	}
}


func (c *Connection) fastForward(localBranch, remoteBranch *git2go.Reference) error {
	_, err := localBranch.SetTarget(remoteBranch.Target(), "Fast-forward")
	if err != nil {
		return fmt.Errorf("error performing fast-forward: %w", err)
	}

	log.Debug().Msg("Fast-forward completed successfully.")
	return nil
}

func (c *Connection) rebase(localBranch, remoteBranch *git2go.Reference) error {
	localAnnotatedCommit, err := c.Repository.AnnotatedCommitFromRef(localBranch)
	if err != nil {
		return fmt.Errorf("error creating annotated commit for local branch: %w", err)
	}

	remoteAnnotatedCommit, err := c.Repository.AnnotatedCommitFromRef(remoteBranch)
	if err != nil {
		return fmt.Errorf("error creating annotated commit for remote branch: %w", err)
	}

	rebase, err := c.Repository.InitRebase(localAnnotatedCommit, remoteAnnotatedCommit, nil, nil)
	if err != nil {
		return fmt.Errorf("error initializing rebase: %w", err)
	}
	defer rebase.Free()

	for {
		operation, err := rebase.Next()
		if err != nil {
			if git2go.IsErrorCode(err, git2go.ErrorCodeIterOver) {
				break
			}
			return fmt.Errorf("error during rebase step: %w", err)
		}

		if operation.Type == git2go.RebaseOperationPick || operation.Type == git2go.RebaseOperationSquash {
			signature := &git2go.Signature{
				Name:  c.Options.Signature.Name,
				Email: c.Options.Signature.Email,
				When:  time.Now(),
			}

			var commitId git2go.Oid
			err = rebase.Commit(&commitId, signature, signature, "rebased")
			if err != nil {
				return fmt.Errorf("error committing during rebase: %w", err)
			}
		}
	}

	index, err := c.Repository.Index()
	if err != nil {
		return fmt.Errorf("error getting repository index: %w", err)
	}

	if index.HasConflicts() {
		return fmt.Errorf("merge conflicts detected; manual resolution required")
	}

	err = rebase.Finish()
	if err != nil {
		return fmt.Errorf("error finishing rebase: %w", err)
	}

	log.Debug().Msg("Rebase completed successfully.")
	return nil
}
