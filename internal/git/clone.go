package git

import (
	"fmt"
	"time"

	"github.com/go-git/go-billy/v5/memfs"
	git "github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/transport"
	"github.com/go-git/go-git/v5/storage/memory"
	log "github.com/sirupsen/logrus"
)

type GitOpsCloneOptions struct {
	Repository string
	Branch     string
	FilePath   string
	Auth       transport.AuthMethod
}

func Clone(options *GitOpsCloneOptions) (*git.Repository, error) {
	startTime := time.Now()
	repo, err := git.Clone(memory.NewStorage(), memfs.New(), &git.CloneOptions{
		URL:           options.Repository,
		Auth:          options.Auth,
		Depth:         1,
		ReferenceName: plumbing.ReferenceName(fmt.Sprintf("refs/heads/%s", options.Branch)),
		SingleBranch:  true,
		Tags:          git.NoTags,
		// Progress:      os.Stdout,
	})
	if err != nil {
		log.Error(err)
		return nil, err
	}
	log.Debug("Cloned repository ", options.Repository, " on branch ", options.Branch, " in ", time.Since(startTime))

	return repo, nil
}
