package git

import (
	"fmt"
	"time"

	git2go "github.com/libgit2/git2go/v34"

	"github.com/rs/zerolog/log"
)

func (c *Connection) Commit(files []string, message string) (string, error) {
	if c.Repository == nil {
		return "", fmt.Errorf("repository is not initialized")
	}

	// Get the index for staging changes
	index, err := c.Repository.Index()
	if err != nil {
		log.Error().Err(err).Msg("Error accessing repository index")
		return "", fmt.Errorf("error accessing repository index: %w", err)
	}

	// Add specified files to the index
	for _, file := range files {
		err = index.AddByPath(file)
		if err != nil {
			log.Error().Err(err).Str("file", file).Msg("Error adding file to index")
			return "", fmt.Errorf("error adding file to index: %w", err)
		}
	}

	// Write the index to the repository's staging area
	err = index.Write()
	if err != nil {
		log.Error().Err(err).Msg("Error writing index to repository")
		return "", fmt.Errorf("error writing index to repository: %w", err)
	}

	treeID, err := index.WriteTree()
	if err != nil {
		log.Error().Err(err).Msg("Error writing tree from index")
		return "", fmt.Errorf("error writing tree from index: %w", err)
	}

	// Lookup the tree object
	tree, err := c.Repository.LookupTree(treeID)
	if err != nil {
		log.Error().Err(err).Msg("Error looking up tree object")
		return "", fmt.Errorf("error looking up tree object: %w", err)
	}

	// Get the HEAD commit for parent
	headRef, err := c.Repository.Head()
	if err != nil {
		log.Error().Err(err).Msg("Error retrieving HEAD reference")
		return "", fmt.Errorf("error retrieving HEAD reference: %w", err)
	}

	var parentCommit *git2go.Commit
	if headRef != nil && headRef.Target() != nil {
		parentCommit, err = c.Repository.LookupCommit(headRef.Target())
		if err != nil {
			log.Error().Err(err).Msg("Error looking up HEAD commit")
			return "", fmt.Errorf("error looking up HEAD commit: %w", err)
		}
	}

	sig := &git2go.Signature{
		Name:  c.Options.Signature.Name,
		Email: c.Options.Signature.Email,
		When:  time.Now(),
	}

	var commitID *git2go.Oid
	if parentCommit != nil {
		commitID, err = c.Repository.CreateCommit(
			"HEAD", sig, sig, message, tree, parentCommit,
		)
	} else {
		commitID, err = c.Repository.CreateCommit(
			"HEAD", sig, sig, message, tree,
		)
	}

	if err != nil {
		log.Error().Err(err).Msg("Error creating commit")
		return "", fmt.Errorf("error creating commit: %w", err)
	}

	log.Debug().Str("commit", commitID.String()).Msg("Commit created successfully")
	return commitID.String(), nil
}
