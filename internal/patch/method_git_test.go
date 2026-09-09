package patch

import (
	"os"
	"path"
	"testing"

	"github.com/google/uuid"
	"github.com/mxcd/gitops-cli/internal/git"
	"github.com/mxcd/gitops-cli/internal/util"
	"github.com/stretchr/testify/assert"
)

func getSshKeyData(t *testing.T) []byte {
	baseDir, err := util.GetGitRepoRoot()
	assert.NoError(t, err)
	assert.NotEmpty(t, baseDir)

	sshKeyPath := path.Join(baseDir, "hack", "soft-serve", "ssh-key")
	assert.FileExists(t, sshKeyPath)

	sshKey, err := os.ReadFile(sshKeyPath)
	assert.NoError(t, err)
	assert.NotEmpty(t, sshKey)

	return sshKey
}

// newTestPatcher clones the soft-serve test repository into a fresh sandbox
// directory and returns a prepared patcher for it.
func newTestPatcher(t *testing.T) *GitPatcher {
	sshKey := getSshKeyData(t)
	baseDir, err := util.GetGitRepoRoot()
	assert.NoError(t, err)

	repositoryPath := path.Join(baseDir, "sandbox", "gitops-test-"+uuid.New().String())
	err = os.MkdirAll(repositoryPath, 0755)
	assert.NoError(t, err)

	gitConnectionOptions := &git.ConnectionOptions{
		Repository:       "ssh://localhost:23231/gitops-test.git",
		Directory:        repositoryPath,
		Branch:           "main",
		IgnoreSshHostKey: true,
		Authentication: &git.Authentication{
			SshKey: &git.SshKey{
				PrivateKey: sshKey,
			},
		},
	}

	patcher, err := NewGitPatcher(&GitPatcherOptions{
		GitConnectionOptions: gitConnectionOptions,
	})
	assert.NoError(t, err)
	assert.NotNil(t, patcher)

	err = patcher.Prepare(&PrepareOptions{Clone: true})
	assert.NoError(t, err)
	assert.NotNil(t, patcher.GitConnection)

	return patcher
}

// seedFixtureFile creates a unique values file in the repository, pushes it and
// returns its repository relative path. Every test works on its own files, so
// tests do not interfere with each other.
func seedFixtureFile(t *testing.T, patcher *GitPatcher) string {
	baseDir, err := util.GetGitRepoRoot()
	assert.NoError(t, err)

	fixtureContents, err := os.ReadFile(path.Join(baseDir, "hack", "soft-serve", "fixtures", "values.yaml"))
	assert.NoError(t, err)
	assert.NotEmpty(t, fixtureContents)

	relativeFilePath := path.Join("applications", "dev", "service-test-"+uuid.New().String(), "values.yaml")
	absoluteFilePath := path.Join(patcher.GitConnection.Options.Directory, relativeFilePath)

	err = os.MkdirAll(path.Dir(absoluteFilePath), 0755)
	assert.NoError(t, err)
	err = os.WriteFile(absoluteFilePath, fixtureContents, 0644)
	assert.NoError(t, err)

	_, err = patcher.GitConnection.Commit([]string{relativeFilePath}, "test: seed "+relativeFilePath)
	assert.NoError(t, err)

	err = patcher.pushWithRetry()
	assert.NoError(t, err)

	return relativeFilePath
}

func readRepositoryFile(t *testing.T, connection *git.Connection, relativeFilePath string) string {
	contents, err := os.ReadFile(path.Join(connection.Options.Directory, relativeFilePath))
	assert.NoError(t, err)
	return string(contents)
}

func TestGitSshPatch(t *testing.T) {
	patcher := newTestPatcher(t)

	patchTask := PatchTask{
		FilePath: "applications/dev/service-test/values.yaml",
		Patches: []Patch{
			{
				Selector: ".service.image.tag",
				Value:    "v1.0.1",
			},
		},
	}

	err := patcher.Patch([]PatchTask{patchTask})
	assert.NoError(t, err)
}

func TestGitSshPatchMissingFile(t *testing.T) {
	patcher := newTestPatcher(t)

	err := patcher.Patch([]PatchTask{{
		FilePath: "applications/does-not-exist/values.yaml",
		Patches:  []Patch{{Selector: ".service.image.tag", Value: "v1.0.1"}},
	}})
	assert.Error(t, err)

	hasChanges, err := patcher.GitConnection.HasChanges()
	assert.NoError(t, err)
	assert.False(t, hasChanges)
}

func TestGitSshPatchBatch(t *testing.T) {
	patcher := newTestPatcher(t)

	relativeFilePathA := seedFixtureFile(t, patcher)
	relativeFilePathB := seedFixtureFile(t, patcher)

	batch := PatchBatch{
		Actor: "ci-bot",
		Files: []FilePatch{
			{FilePath: relativeFilePathA, Patches: []Patch{{Selector: ".service.image.tag", Value: "v2.0.0"}}},
			{FilePath: relativeFilePathB, Patches: []Patch{{Selector: ".service.global.namespace", Value: "batch-namespace"}}},
		},
	}

	commitHash, err := patcher.PatchBatch(batch)
	assert.NoError(t, err)
	assert.NotEmpty(t, commitHash)

	// both files must be part of the very same commit
	committedFiles, err := patcher.GitConnection.CommitFiles(commitHash)
	assert.NoError(t, err)
	assert.ElementsMatch(t, []string{relativeFilePathA, relativeFilePathB}, committedFiles)

	// the changes must be visible in an independent clone of the repository
	verificationPatcher := newTestPatcher(t)
	resolvedCommitHash, err := verificationPatcher.GitConnection.RevParse(commitHash)
	assert.NoError(t, err)
	assert.Equal(t, commitHash, resolvedCommitHash)

	assert.Contains(t, readRepositoryFile(t, verificationPatcher.GitConnection, relativeFilePathA), "tag: v2.0.0")
	assert.Contains(t, readRepositoryFile(t, verificationPatcher.GitConnection, relativeFilePathB), "namespace: batch-namespace")
}

func TestGitSshPatchBatchSingleFile(t *testing.T) {
	patcher := newTestPatcher(t)

	relativeFilePath := seedFixtureFile(t, patcher)

	commitHash, err := patcher.PatchBatch(PatchBatch{
		Files: []FilePatch{
			{FilePath: relativeFilePath, Patches: []Patch{{Selector: ".service.image.tag", Value: "v3.0.0"}}},
		},
	})
	assert.NoError(t, err)
	assert.NotEmpty(t, commitHash)

	committedFiles, err := patcher.GitConnection.CommitFiles(commitHash)
	assert.NoError(t, err)
	assert.Equal(t, []string{relativeFilePath}, committedFiles)
}

func TestGitSshPatchBatchNoChanges(t *testing.T) {
	patcher := newTestPatcher(t)

	relativeFilePath := seedFixtureFile(t, patcher)

	batch := PatchBatch{
		Files: []FilePatch{
			{FilePath: relativeFilePath, Patches: []Patch{{Selector: ".service.image.tag", Value: "v4.0.0"}}},
		},
	}

	commitHash, err := patcher.PatchBatch(batch)
	assert.NoError(t, err)
	assert.NotEmpty(t, commitHash)

	// applying the same batch again must not create another commit
	commitHash, err = patcher.PatchBatch(batch)
	assert.NoError(t, err)
	assert.Empty(t, commitHash)
}

func TestGitSshPatchBatchMissingFile(t *testing.T) {
	patcher := newTestPatcher(t)

	relativeFilePath := seedFixtureFile(t, patcher)
	originalContents := readRepositoryFile(t, patcher.GitConnection, relativeFilePath)

	commitHash, err := patcher.PatchBatch(PatchBatch{
		Files: []FilePatch{
			{FilePath: relativeFilePath, Patches: []Patch{{Selector: ".service.image.tag", Value: "v5.0.0"}}},
			{FilePath: "applications/does-not-exist/values.yaml", Patches: []Patch{{Selector: ".service.image.tag", Value: "v5.0.0"}}},
		},
	})
	assert.Error(t, err)
	assert.Empty(t, commitHash)

	// the valid file must not have been touched
	assert.Equal(t, originalContents, readRepositoryFile(t, patcher.GitConnection, relativeFilePath))

	hasChanges, err := patcher.GitConnection.HasChanges()
	assert.NoError(t, err)
	assert.False(t, hasChanges)
}

func TestGitSshPatchBatchInvalidSelector(t *testing.T) {
	patcher := newTestPatcher(t)

	relativeFilePathA := seedFixtureFile(t, patcher)
	relativeFilePathB := seedFixtureFile(t, patcher)
	originalContents := readRepositoryFile(t, patcher.GitConnection, relativeFilePathA)

	commitHash, err := patcher.PatchBatch(PatchBatch{
		Files: []FilePatch{
			{FilePath: relativeFilePathA, Patches: []Patch{{Selector: ".service.image.tag", Value: "v6.0.0"}}},
			{FilePath: relativeFilePathB, Patches: []Patch{{Selector: ".does.not.exist", Value: "v6.0.0"}}},
		},
	})
	assert.Error(t, err)
	assert.Empty(t, commitHash)

	// the failure happens while patching in memory, so no file is written
	assert.Equal(t, originalContents, readRepositoryFile(t, patcher.GitConnection, relativeFilePathA))

	hasChanges, err := patcher.GitConnection.HasChanges()
	assert.NoError(t, err)
	assert.False(t, hasChanges)
}

func TestGitSshPatchBatchInvalidFilePath(t *testing.T) {
	patcher := newTestPatcher(t)

	commitHash, err := patcher.PatchBatch(PatchBatch{
		Files: []FilePatch{
			{FilePath: "../escape.yaml", Patches: []Patch{{Selector: ".service.image.tag", Value: "v7.0.0"}}},
		},
	})
	assert.Error(t, err)
	assert.ErrorIs(t, err, ErrInvalidPatchBatch)
	assert.Empty(t, commitHash)
}

func TestGitSshPatchBatchWriteFailureRestoresWorkingTree(t *testing.T) {
	if os.Geteuid() == 0 {
		t.Skip("read only files are still writable as root")
	}

	patcher := newTestPatcher(t)

	relativeFilePathA := seedFixtureFile(t, patcher)
	relativeFilePathB := seedFixtureFile(t, patcher)
	originalContents := readRepositoryFile(t, patcher.GitConnection, relativeFilePathA)

	// make the second file read only so that the write phase fails after the
	// first file has already been written
	absoluteFilePathB := path.Join(patcher.GitConnection.Options.Directory, relativeFilePathB)
	err := os.Chmod(absoluteFilePathB, 0444)
	assert.NoError(t, err)
	defer func() {
		assert.NoError(t, os.Chmod(absoluteFilePathB, 0644))
	}()

	commitHash, err := patcher.PatchBatch(PatchBatch{
		Files: []FilePatch{
			{FilePath: relativeFilePathA, Patches: []Patch{{Selector: ".service.image.tag", Value: "v8.0.0"}}},
			{FilePath: relativeFilePathB, Patches: []Patch{{Selector: ".service.image.tag", Value: "v8.0.0"}}},
		},
	})
	assert.Error(t, err)
	assert.Empty(t, commitHash)

	// the already written file must have been restored
	assert.Equal(t, originalContents, readRepositoryFile(t, patcher.GitConnection, relativeFilePathA))

	hasChanges, err := patcher.GitConnection.HasChanges()
	assert.NoError(t, err)
	assert.False(t, hasChanges)
}
