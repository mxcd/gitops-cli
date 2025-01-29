package git

import (
	git2go "github.com/libgit2/git2go/v34"
)

type FetchOptionsUsage string

const (
	FetchOptionUsageDefault FetchOptionsUsage = "default"
)

func (c *Connection) getFetchOptions(usage FetchOptionsUsage) git2go.FetchOptions {

	fetchOptions := git2go.FetchOptions{}

	switch usage {
	case FetchOptionUsageDefault:
		fetchOptions.RemoteCallbacks = git2go.RemoteCallbacks{
			CredentialsCallback: c.credentialsCallback,
		}
	}

	if c.Options.IgnoreSslHostKey {
		fetchOptions.RemoteCallbacks.CertificateCheckCallback = func(cert *git2go.Certificate, valid bool, hostname string) error {
			return nil
		}
	}

	return fetchOptions
}

// // mustAnnotatedCommit is a helper to create an annotated commit.
// func mustAnnotatedCommit(repo *git2go.Repository, id git2go.Oid) *git2go.AnnotatedCommit {
// 	ann, err := repo.AnnotatedCommitLookup(&id)
// 	if err != nil {
// 		panic(err)
// 	}
// 	return ann
// }

// // Commit stages the given files and creates a new commit.
// // Returns the new commit OID on success.
// func (c *Connection) Commit(files []string, message string) (*git2go.Oid, error) {
// 	if c.Repository == nil {
// 		return nil, errors.New("repository is nil, clone or open a repo first")
// 	}

// 	index, err := c.Repository.Index()
// 	if err != nil {
// 		log.Error().Err(err).Msg("Error getting index")
// 		return nil, err
// 	}

// 	for _, file := range files {
// 		err := index.AddByPath(file)
// 		if err != nil {
// 			log.Error().Err(err).Msg("Error adding file to index")
// 			return nil, err
// 		}
// 	}

// 	if err := index.Write(); err != nil {
// 		log.Error().Err(err).Msg("Error writing index")
// 		return nil, err
// 	}

// 	treeID, err := index.WriteTree()
// 	if err != nil {
// 		log.Error().Err(err).Msg("Error writing tree")
// 		return nil, err
// 	}

// 	tree, err := c.Repository.LookupTree(treeID)
// 	if err != nil {
// 		log.Error().Err(err).Msg("Error looking up tree")
// 		return nil, err
// 	}

// 	sig := &git2go.Signature{
// 		Name:  "git2go-user",
// 		Email: "example@domain",
// 	}

// 	head, err := c.Repository.Head()
// 	if err != nil {
// 		log.Error().Err(err).Msg("Error getting HEAD")
// 		return nil, err
// 	}

// 	parent, err := c.Repository.LookupCommit(head.Target())
// 	if err != nil {
// 		log.Error().Err(err).Msg("Error looking up HEAD commit")
// 		return nil, err
// 	}

// 	commitID, err := c.Repository.CreateCommit(head.Name(), sig, sig, message, tree, parent)
// 	if err != nil {
// 		log.Error().Err(err).Msg("Error creating commit")
// 		return nil, err
// 	}

// 	return commitID, nil
// }

// // Push attempts to push the given branch to origin.
// func (c *Connection) Push(branch string) error {
// 	if c.Repository == nil {
// 		return errors.New("repository is nil, clone or open a repo first")
// 	}

// 	remote, err := c.Repository.Remotes.Lookup("origin")
// 	if err != nil {
// 		log.Error().Err(err).Msg("Error looking up remote 'origin'")
// 		return err
// 	}

// 	refSpec := fmt.Sprintf("refs/heads/%s:refs/heads/%s", branch, branch)

// 	pushOpts := &git2go.PushOptions{
// 		RemoteCallbacks: git2go.RemoteCallbacks{
// 			CredentialsCallback: c.credentialsCallback,
// 		},
// 	}

// 	err = remote.Push([]string{refSpec}, pushOpts)
// 	if err != nil {
// 		log.Error().Err(err).Msgf("Error pushing to branch %s", branch)
// 		return err
// 	}

// 	log.Info().Msgf("Pushed to origin/%s", branch)
// 	return nil
// }

// // HasChanges checks whether the index differs from HEAD.
// func (c *Connection) HasChanges() (bool, error) {
// 	if c.Repository == nil {
// 		return false, errors.New("repository is nil, clone or open a repo first")
// 	}

// 	head, err := c.Repository.Head()
// 	if err != nil {
// 		log.Error().Err(err).Msg("Error getting HEAD")
// 		return false, err
// 	}

// 	headCommit, err := c.Repository.LookupCommit(head.Target())
// 	if err != nil {
// 		log.Error().Err(err).Msg("Error looking up HEAD commit")
// 		return false, err
// 	}

// 	headTree, err := headCommit.Tree()
// 	if err != nil {
// 		log.Error().Err(err).Msg("Error getting HEAD tree")
// 		return false, err
// 	}

// 	index, err := c.Repository.Index()
// 	if err != nil {
// 		return false, err
// 	}

// 	diff, err := c.Repository.DiffTreeToIndex(headTree, index, nil)
// 	if err != nil {
// 		return false, err
// 	}

// 	return diff.NumDeltas() > 0, nil
// }
