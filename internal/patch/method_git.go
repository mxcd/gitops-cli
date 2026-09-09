package patch

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"time"

	"github.com/mxcd/gitops-cli/internal/git"
	"github.com/mxcd/gitops-cli/internal/yaml"
	"github.com/urfave/cli/v2"

	"github.com/rs/zerolog/log"
)

// pushRetryCount is the number of pull/push attempts before giving up.
const pushRetryCount = 5

type GitPatcherOptions struct {
	GitConnectionOptions *git.ConnectionOptions
	GitConnection        *git.Connection
}

type GitPatcher struct {
	Options       *GitPatcherOptions
	GitConnection *git.Connection
}

func GetGitConnectionOptionsFromCli(c *cli.Context) (*git.ConnectionOptions, error) {

	var authentication *git.Authentication = nil

	if c.String("basicauth") != "" {
		auth, err := git.GetAuthFromBasicAuthString(c.String("basicauth"))
		if err != nil {
			return nil, err
		}
		authentication = auth
	} else if c.String("ssh-key") != "" || c.String("ssh-key-file") != "" {
		var sshKey []byte
		if c.String("ssh-key") != "" {
			sshKey = []byte(c.String("ssh-key"))
		} else {
			file, err := os.ReadFile(c.String("ssh-key-file"))
			if err != nil {
				return nil, err
			}
			sshKey = file
		}

		var passphrase *string
		if c.String("ssh-key-passphrase") != "" {
			_passphrase := c.String("ssh-key-passphrase")
			passphrase = &_passphrase
		}

		auth, err := git.GetAuthFromSshKey(sshKey, passphrase)
		if err != nil {
			return nil, err
		}
		authentication = auth
	}

	options := &git.ConnectionOptions{
		Repository:     c.String("repository"),
		Branch:         c.String("branch"),
		Authentication: authentication,
	}

	return options, nil
}

func GetGitPatcherOptionsFromCli(c *cli.Context) (*GitPatcherOptions, error) {

	gitConnectionOptions, err := GetGitConnectionOptionsFromCli(c)
	if err != nil {
		return nil, err
	}

	options := &GitPatcherOptions{
		GitConnectionOptions: gitConnectionOptions,
	}

	return options, nil
}

func NewGitPatcher(options *GitPatcherOptions) (*GitPatcher, error) {
	return &GitPatcher{
		Options: options,
	}, nil
}

func (p *GitPatcher) Prepare(options *PrepareOptions) error {

	if p.Options.GitConnection != nil {
		p.GitConnection = p.Options.GitConnection
	} else {
		gitConnection, err := git.NewGitConnection(p.Options.GitConnectionOptions)
		if err != nil {
			return err
		}
		p.GitConnection = gitConnection
	}

	if options != nil && options.Clone {
		err := p.GitConnection.Clone()
		if err != nil {
			return err
		}
	}

	return nil
}

// patchedFile holds the fully patched contents of a single file before it is
// written to disk.
type patchedFile struct {
	RelativePath string
	AbsolutePath string
	Mode         fs.FileMode
	Contents     []byte
}

// preparePatchedFile reads a file and applies all of its patches in memory. No
// changes are written to disk, so a failing file never leaves a partially
// patched working tree behind.
func (p *GitPatcher) preparePatchedFile(file FilePatch) (*patchedFile, error) {
	relativeFilePath, err := cleanRelativeFilePath(file.FilePath)
	if err != nil {
		return nil, fmt.Errorf("%w: filePath '%s': %s", ErrInvalidPatchBatch, file.FilePath, err.Error())
	}

	absoluteFilePath := filepath.Join(p.GitConnection.Options.Directory, relativeFilePath)

	// Lstat, so a symlink is seen as such and cannot redirect the write to
	// another file inside or outside of the repository
	fileStat, err := os.Lstat(absoluteFilePath)
	if err != nil {
		log.Error().Err(err).Msgf("Failed to stat file %s", relativeFilePath)
		return nil, fmt.Errorf("failed to stat file %s: %w", relativeFilePath, err)
	}
	if !fileStat.Mode().IsRegular() {
		return nil, fmt.Errorf("%w: filePath '%s': is not a regular file", ErrInvalidPatchBatch, file.FilePath)
	}

	fileContents, err := os.ReadFile(absoluteFilePath)
	if err != nil {
		log.Error().Err(err).Msgf("Failed to read file %s", relativeFilePath)
		return nil, fmt.Errorf("failed to read file %s: %w", relativeFilePath, err)
	}

	log.Debug().Msgf("original yaml file: %s", string(fileContents))

	for _, filePatch := range file.Patches {
		selector := filePatch.Selector
		value := filePatch.Value

		log.Debug().Msgf("patching file '%s' with selector '%s' and value '%s'", relativeFilePath, selector, value)
		patchedData, err := yaml.PatchYaml(fileContents, selector, value)
		if err != nil {
			return nil, fmt.Errorf("failed to patch file %s with selector '%s': %w", relativeFilePath, selector, err)
		}
		log.Debug().Msgf("patched yaml file:\n%s", string(patchedData))

		fileContents = patchedData
	}

	return &patchedFile{
		RelativePath: relativeFilePath,
		AbsolutePath: absoluteFilePath,
		Mode:         fileStat.Mode(),
		Contents:     fileContents,
	}, nil
}

// writePatchedFile writes the patched contents back to disk.
func (p *GitPatcher) writePatchedFile(file *patchedFile) error {
	if err := os.WriteFile(file.AbsolutePath, file.Contents, file.Mode); err != nil {
		log.Error().Err(err).Msgf("Failed to write file %s", file.RelativePath)
		return fmt.Errorf("failed to write file %s: %w", file.RelativePath, err)
	}
	return nil
}

// pushWithRetry pulls and pushes with a linear backoff to resolve races with
// concurrent writers to the repository. It returns the id of HEAD after the
// push, which differs from the local commit id if the pull rebased it. A
// failing pull is not retried: it means the rebase conflicts with upstream, and
// waiting does not resolve that.
func (p *GitPatcher) pushWithRetry() (string, error) {
	var err error
	for i := 0; i < pushRetryCount; i++ {
		if err = p.GitConnection.Pull(); err != nil {
			log.Error().Err(err).Msg("Error pulling prior to push")
			return "", err
		}
		if err = p.GitConnection.Push(); err == nil {
			return p.GitConnection.RevParse("HEAD")
		}
		if i < pushRetryCount-1 {
			time.Sleep(time.Duration(i+1) * time.Second)
		}
	}

	return "", err
}

// buildCommitMessage builds the commit message for the given repository
// relative file paths.
func buildCommitMessage(relativeFilePaths []string, actor string) string {
	var message string
	if len(relativeFilePaths) == 1 {
		message = fmt.Sprintf("feat(gitops): patching %s", relativeFilePaths[0])
	} else {
		message = fmt.Sprintf("feat(gitops): patching %d files\n", len(relativeFilePaths))
		for _, relativeFilePath := range relativeFilePaths {
			message += fmt.Sprintf("\n- %s", relativeFilePath)
		}
	}

	if actor != "" {
		message += fmt.Sprintf("\n\nTriggered by: %s", actor)
	}

	return message
}

// Patch applies the given tasks as a single batch, see PatchBatch.
func (p *GitPatcher) Patch(patchTasks []PatchTask) error {
	batch := PatchBatch{Files: make([]FilePatch, 0, len(patchTasks))}
	for _, patchTask := range patchTasks {
		if batch.Actor == "" {
			batch.Actor = patchTask.Actor
		}
		batch.Files = append(batch.Files, FilePatch{FilePath: patchTask.FilePath, Patches: patchTask.Patches})
	}

	_, err := p.PatchBatch(batch)
	return err
}

// PatchBatch applies all patches of the batch and commits them as a single
// atomic commit. It returns the commit id, or an empty string if the batch did
// not change anything. Any failure resets the clone to the remote branch, so
// nothing is ever left half written, staged or committed but unpushed.
func (p *GitPatcher) PatchBatch(batch PatchBatch) (hash string, err error) {

	if err := ValidatePatchBatch(batch); err != nil {
		return "", err
	}

	defer func() {
		if err != nil {
			log.Warn().Msg("Resetting clone to the remote branch after failed batch patch")
			if resetErr := p.GitConnection.ResetToUpstream(); resetErr != nil {
				log.Error().Err(resetErr).Msg("Failed to reset clone after failed batch patch")
			}
		}
	}()

	if err = p.GitConnection.Pull(); err != nil {
		return "", err
	}

	// patch all files in memory first, so that a failure of any file does not
	// leave the working tree with a partially applied batch
	preparedFiles := make([]*patchedFile, 0, len(batch.Files))
	relativeFilePaths := make([]string, 0, len(batch.Files))
	for _, file := range batch.Files {
		preparedFile, err := p.preparePatchedFile(file)
		if err != nil {
			return "", err
		}
		preparedFiles = append(preparedFiles, preparedFile)
		relativeFilePaths = append(relativeFilePaths, preparedFile.RelativePath)
	}

	// an untracked file would be written but never committed
	if err = p.GitConnection.RequireTracked(relativeFilePaths); err != nil {
		return "", err
	}

	for _, preparedFile := range preparedFiles {
		if err = p.writePatchedFile(preparedFile); err != nil {
			return "", err
		}
	}

	log.Debug().Msg("checking for changes")
	hasChanges, err := p.GitConnection.HasChanges()
	if err != nil {
		return "", err
	}

	if !hasChanges {
		log.Info().Msg("No changes detected, nothing to commit")
		return "", nil
	}

	log.Debug().Msg("Changes detected, committing")

	localHash, err := p.GitConnection.Commit(relativeFilePaths, buildCommitMessage(relativeFilePaths, batch.Actor))
	if err != nil {
		return "", err
	}
	log.Info().Msgf("Created patch commit: %s", localHash)

	hash, err = p.pushWithRetry()
	if err != nil {
		log.Error().Err(err).Msgf("Failed to push commit %s, discarding it", localHash)
		return "", err
	}

	return hash, nil
}
