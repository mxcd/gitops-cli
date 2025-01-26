package git

import (
	"fmt"
	"time"

	"github.com/go-git/go-billy/v5/memfs"
	git "github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/storage/memory"
	log "github.com/sirupsen/logrus"
)

func (c *GitConnection) Clone() error {
	startTime := time.Now()
	repo, err := git.Clone(memory.NewStorage(), memfs.New(), &git.CloneOptions{
		URL:           c.Options.Repository,
		Auth:          c.Options.Auth,
		Depth:         1,
		ReferenceName: plumbing.ReferenceName(fmt.Sprintf("refs/heads/%s", c.Options.Branch)),
		SingleBranch:  true,
		Tags:          git.NoTags,
		// Progress:      os.Stdout,
	})
	if err != nil {
		log.WithError(err).Error("Error cloning repository ", c.Options.Repository, " on branch ", c.Options.Branch)
		return err
	}
	log.Debug("Cloned repository ", c.Options.Repository, " on branch ", c.Options.Branch, " in ", time.Since(startTime))

	c.Repository = repo

	return nil
}

func (c *GitConnection) Commit(files []string, message string) (*plumbing.Hash, error) {
	worktree, err := c.Repository.Worktree()
	if err != nil {
		log.WithError(err).Error("Error getting worktree")
		return nil, err
	}

	for _, file := range files {
		_, err := worktree.Add(file)
		if err != nil {
			log.WithError(err).Error("Error 'git add'ing file")
			return nil, err
		}
	}

	hash, err := worktree.Commit(message, &git.CommitOptions{})
	if err != nil {
		log.WithError(err).Error("Error 'git commit'ing files")
	}

	return &hash, err
}

func (c *GitConnection) Push(branch string) error {
	err := c.Repository.Push(&git.PushOptions{
		RemoteName: "origin",
		Auth:       c.Options.Auth,
	})
	if err != nil {
		log.WithError(err).Error("Error pushing to branch ", branch)
		return err
	}

	log.Info("Pushed to origin/", branch)

	return nil
}

func (c *GitConnection) HasChanges() (bool, error) {
	worktree, err := c.Repository.Worktree()
	if err != nil {
		log.WithError(err).Error("Error getting worktree")
		return false, err
	}

	status, err := worktree.Status()
	if err != nil {
		log.WithError(err).Error("Error getting status")
		return false, err
	}

	return !status.IsClean(), nil
}
