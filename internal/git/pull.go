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

	isAncestor, err := c.Repository.DescendantOf(remoteBranch.Target(), localBranch.Target())
	if err != nil {
		return fmt.Errorf("error checking if fast-forward is possible: %w", err)
	}

	log.Debug().Bool("isAncestor", isAncestor).Msg("Checking if fast-forward is possible")

	if isAncestor {
		return c.fastForward(localBranch, remoteBranch)
	} else if c.Options.PullRebase {
		return c.rebase(localBranch, remoteBranch)
	} else {
		return c.merge(remoteBranch)
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

func (c *Connection) merge(remoteBranch *git2go.Reference) error {
	remoteAnnotatedCommit, err := c.Repository.AnnotatedCommitFromRef(remoteBranch)
	if err != nil {
		return fmt.Errorf("error creating annotated commit for remote branch: %w", err)
	}

	err = c.Repository.Merge([]*git2go.AnnotatedCommit{remoteAnnotatedCommit}, nil, nil)
	if err != nil {
		return fmt.Errorf("error performing merge: %w", err)
	}

	index, err := c.Repository.Index()
	if err != nil {
		return fmt.Errorf("error getting repository index: %w", err)
	}

	if index.HasConflicts() {
		return fmt.Errorf("merge conflicts detected; manual resolution required")
	}

	sig := &git2go.Signature{
		Name:  c.Options.Signature.Name,
		Email: c.Options.Signature.Email,
		When:  time.Now(),
	}

	treeID, err := index.WriteTree()
	if err != nil {
		return fmt.Errorf("error writing tree from index: %w", err)
	}

	tree, err := c.Repository.LookupTree(treeID)
	if err != nil {
		return fmt.Errorf("error looking up tree: %w", err)
	}

	headRef, err := c.Repository.Head()
	if err != nil {
		return fmt.Errorf("error getting HEAD reference: %w", err)
	}

	parentCommit, err := c.Repository.LookupCommit(headRef.Target())
	if err != nil {
		return fmt.Errorf("error looking up HEAD commit: %w", err)
	}

	_, err = c.Repository.CreateCommit("HEAD", sig, sig, "Merge branch '"+c.Options.Branch+"' into HEAD", tree, parentCommit)
	if err != nil {
		return fmt.Errorf("error creating merge commit: %w", err)
	}

	err = c.Repository.StateCleanup()
	if err != nil {
		return fmt.Errorf("error cleaning up merge state: %w", err)
	}

	log.Debug().Msg("Merge completed successfully.")
	return nil
}
