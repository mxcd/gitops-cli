package git

import (
	"log"
	"os"
	"path"
	"testing"

	"github.com/google/uuid"
	"github.com/ldez/go-git-cmd-wrapper/v2/add"
	"github.com/ldez/go-git-cmd-wrapper/v2/git"
	"github.com/ldez/go-git-cmd-wrapper/v2/types"
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

func TestNewGitConnection(t *testing.T) {

	sshKey := getSshKeyData(t)

	baseDir, err := util.GetGitRepoRoot()
	assert.NoError(t, err)

	authentication, err := GetAuthFromSshKey(sshKey, nil)
	assert.NoError(t, err)
	assert.NotNil(t, authentication)

	uuid := uuid.New().String()

	options := &ConnectionOptions{
		Directory:        path.Join(baseDir, "sandbox", "gitops-test-"+uuid),
		Repository:       "ssh://git@localhost:23231/gitops-test.git",
		Branch:           "main",
		Authentication:   authentication,
		IgnoreSshHostKey: true,
	}
	gitConnection, err := NewGitConnection(options)
	assert.NoError(t, err)
	assert.NotNil(t, gitConnection)

	err = gitConnection.Clone()
	assert.NoError(t, err)
}

func cloneTempRepository(t *testing.T) *Connection {
	sshKey := getSshKeyData(t)

	baseDir, err := util.GetGitRepoRoot()
	assert.NoError(t, err)

	uuid := uuid.New().String()
	directoryName := path.Join(baseDir, "sandbox", "gitops-test-"+uuid)

	authentication, err := GetAuthFromSshKey(sshKey, nil)
	assert.NoError(t, err)
	assert.NotNil(t, authentication)

	options := &ConnectionOptions{
		Directory:        directoryName,
		Repository:       "ssh://git@localhost:23231/gitops-test.git",
		Branch:           "main",
		Authentication:   authentication,
		IgnoreSshHostKey: true,
	}
	gitConnection, err := NewGitConnection(options)
	assert.NoError(t, err)
	assert.NotNil(t, gitConnection)

	err = gitConnection.Clone()
	assert.NoError(t, err)

	return gitConnection
}

func TestGitPullFastForward(t *testing.T) {

	tempConnectionA := cloneTempRepository(t)
	assert.NotNil(t, tempConnectionA)
	log.Println("tempConnectionA cloned")

	tempConnectionB := cloneTempRepository(t)
	assert.NotNil(t, tempConnectionB)
	log.Println("tempConnectionB cloned")

	uuid := uuid.New().String()
	testFileName := "test-file-" + uuid
	testFilePath := path.Join(tempConnectionA.Options.Directory, testFileName)
	err := os.WriteFile(testFilePath, []byte("test"), 0644)
	assert.NoError(t, err)

	log.Printf("file written to %s", testFilePath)

	hash, err := tempConnectionA.Commit([]string{testFileName}, "Test commit")
	assert.NoError(t, err)
	assert.NotEmpty(t, hash)

	err = tempConnectionA.Push()
	assert.NoError(t, err)

	err = tempConnectionB.Pull()
	assert.NoError(t, err)

	testFilePath = path.Join(tempConnectionB.Options.Directory, testFileName)
	_, err = os.Stat(testFilePath)
	assert.NoError(t, err)
	data, err := os.ReadFile(testFilePath)
	assert.NoError(t, err)
	assert.Equal(t, "test", string(data))
}

func TestGitPullRebase(t *testing.T) {

	// Clone repository A
	tempConnectionA := cloneTempRepository(t)
	assert.NotNil(t, tempConnectionA)
	log.Println("tempConnectionA cloned")

	// Clone repository B
	tempConnectionB := cloneTempRepository(t)
	assert.NotNil(t, tempConnectionB)
	log.Println("tempConnectionB cloned")

	// Create a new file in repository A, commit and push it
	uuidA := uuid.New().String()
	testFileNameA := "test-file-" + uuidA
	testFilePathA := path.Join(tempConnectionA.Options.Directory, testFileNameA)
	err := os.WriteFile(testFilePathA, []byte("test A"), 0644)
	assert.NoError(t, err)

	log.Printf("file written to %s", testFilePathA)

	hashA, err := tempConnectionA.Commit([]string{testFileNameA}, "Test commit A")
	assert.NoError(t, err)
	assert.NotEmpty(t, hashA)

	err = tempConnectionA.Push()
	assert.NoError(t, err)

	// Create a new file in repository B, commit and push it
	uuidB := uuid.New().String()
	testFileNameB := "test-file-" + uuidB
	testFilePathB := path.Join(tempConnectionB.Options.Directory, testFileNameB)
	err = os.WriteFile(testFilePathB, []byte("test B"), 0644)
	assert.NoError(t, err)

	log.Printf("file written to %s", testFilePathB)

	hashB, err := tempConnectionB.Commit([]string{testFileNameB}, "Test commit B")
	assert.NoError(t, err)
	assert.NotEmpty(t, hashB)

	// Push is expected to fail because of changes from repository A in remote
	err = tempConnectionB.Push()
	assert.Error(t, err)

	// Pull is expected to rebase the changes from repository A
	err = tempConnectionB.Pull()
	assert.NoError(t, err)

	// check if test file A is added to repo B
	testFilePath := path.Join(tempConnectionB.Options.Directory, testFileNameA)
	_, err = os.Stat(testFilePath)
	assert.NoError(t, err)
	data, err := os.ReadFile(testFilePath)
	assert.NoError(t, err)
	assert.Equal(t, "test A", string(data))

	// Push is expected to succeed after rebase
	err = tempConnectionB.Push()
	assert.NoError(t, err)

	// Pull is expected to pull the changes from repository B
	err = tempConnectionA.Pull()
	assert.NoError(t, err)

	// check if test file B is added to repo A
	testFilePath = path.Join(tempConnectionA.Options.Directory, testFileNameB)
	_, err = os.Stat(testFilePath)
	assert.NoError(t, err)
	data, err = os.ReadFile(testFilePath)
	assert.NoError(t, err)
	assert.Equal(t, "test B", string(data))
}

func TestGitResetToUpstream(t *testing.T) {

	tempConnection := cloneTempRepository(t)
	assert.NotNil(t, tempConnection)

	relativeFilePath := path.Join("applications", "dev", "service-test", "values.yaml")
	absoluteFilePath := path.Join(tempConnection.Options.Directory, relativeFilePath)

	upstreamHead, err := tempConnection.RevParse("origin/main")
	assert.NoError(t, err)

	// a local commit that was never pushed, plus a staged and an unstaged change on top
	err = os.WriteFile(absoluteFilePath, []byte("clobbered: true\n"), 0644)
	assert.NoError(t, err)
	localCommit, err := tempConnection.Commit([]string{relativeFilePath}, "local only")
	assert.NoError(t, err)
	assert.NotEqual(t, upstreamHead, localCommit)

	err = os.WriteFile(absoluteFilePath, []byte("clobbered: staged\n"), 0644)
	assert.NoError(t, err)
	_, err = git.Add(runGitIn(tempConnection.Options.Directory), add.PathSpec(relativeFilePath))
	assert.NoError(t, err)
	err = os.WriteFile(absoluteFilePath, []byte("clobbered: unstaged\n"), 0644)
	assert.NoError(t, err)
	assert.True(t, hasStagedChanges(t, tempConnection))

	err = tempConnection.ResetToUpstream()
	assert.NoError(t, err)

	head, err := tempConnection.RevParse("HEAD")
	assert.NoError(t, err)
	assert.Equal(t, upstreamHead, head)
	assert.False(t, hasStagedChanges(t, tempConnection))
	hasChanges, err := tempConnection.HasChanges()
	assert.NoError(t, err)
	assert.False(t, hasChanges)
}

func TestGitResetToUpstreamAbortsRebase(t *testing.T) {

	tempConnection := cloneTempRepository(t)
	otherConnection := cloneTempRepository(t)

	// a file of its own, so the shared fixture file stays untouched
	relativeFilePath := "conflict-file-" + uuid.New().String()
	err := os.WriteFile(path.Join(otherConnection.Options.Directory, relativeFilePath), []byte("conflict: none\n"), 0644)
	assert.NoError(t, err)
	_, err = otherConnection.Commit([]string{relativeFilePath}, "seed conflict file")
	assert.NoError(t, err)
	assert.NoError(t, otherConnection.Push())
	assert.NoError(t, tempConnection.Pull())

	// upstream and the local clone change the same line
	err = os.WriteFile(path.Join(otherConnection.Options.Directory, relativeFilePath), []byte("conflict: upstream\n"), 0644)
	assert.NoError(t, err)
	_, err = otherConnection.Commit([]string{relativeFilePath}, "upstream change")
	assert.NoError(t, err)
	assert.NoError(t, otherConnection.Push())

	err = os.WriteFile(path.Join(tempConnection.Options.Directory, relativeFilePath), []byte("conflict: local\n"), 0644)
	assert.NoError(t, err)
	_, err = tempConnection.Commit([]string{relativeFilePath}, "local change")
	assert.NoError(t, err)

	// the rebase conflicts and leaves the clone mid-rebase
	assert.Error(t, tempConnection.Pull())
	assert.DirExists(t, path.Join(tempConnection.Options.Directory, ".git", "rebase-merge"))

	assert.NoError(t, tempConnection.ResetToUpstream())

	assert.NoDirExists(t, path.Join(tempConnection.Options.Directory, ".git", "rebase-merge"))
	head, err := tempConnection.RevParse("HEAD")
	assert.NoError(t, err)
	upstreamHead, err := tempConnection.RevParse("origin/main")
	assert.NoError(t, err)
	assert.Equal(t, upstreamHead, head)
	assert.NoError(t, tempConnection.Pull())
}

func TestGitRequireTracked(t *testing.T) {

	tempConnection := cloneTempRepository(t)

	tracked := path.Join("applications", "dev", "service-test", "values.yaml")
	untracked := path.Join("applications", "dev", "service-test", "untracked.yaml")
	err := os.WriteFile(path.Join(tempConnection.Options.Directory, untracked), []byte("untracked: true\n"), 0644)
	assert.NoError(t, err)

	assert.NoError(t, tempConnection.RequireTracked([]string{tracked}))
	assert.Error(t, tempConnection.RequireTracked([]string{untracked}))
	assert.Error(t, tempConnection.RequireTracked([]string{tracked, untracked}))
}

// hasStagedChanges reports whether the index differs from HEAD.
func hasStagedChanges(t *testing.T, connection *Connection) bool {
	_, err := git.Raw("diff", runGitIn(connection.Options.Directory), func(g *types.Cmd) {
		g.AddOptions("--cached")
		g.AddOptions("--quiet")
	})
	return err != nil
}

func TestGitCommitFiles(t *testing.T) {

	tempConnection := cloneTempRepository(t)
	assert.NotNil(t, tempConnection)

	headBefore, err := tempConnection.RevParse("HEAD")
	assert.NoError(t, err)
	assert.NotEmpty(t, headBefore)

	testFileNameA := "test-file-" + uuid.New().String()
	testFileNameB := "test-file-" + uuid.New().String()

	for _, testFileName := range []string{testFileNameA, testFileNameB} {
		err = os.WriteFile(path.Join(tempConnection.Options.Directory, testFileName), []byte("test"), 0644)
		assert.NoError(t, err)
	}

	hash, err := tempConnection.Commit([]string{testFileNameA, testFileNameB}, "Test commit with two files")
	assert.NoError(t, err)
	assert.NotEmpty(t, hash)

	files, err := tempConnection.CommitFiles(hash)
	assert.NoError(t, err)
	assert.ElementsMatch(t, []string{testFileNameA, testFileNameB}, files)

	// the commit must sit directly on top of the previously resolved HEAD
	parent, err := tempConnection.RevParse(hash + "^")
	assert.NoError(t, err)
	assert.Equal(t, headBefore, parent)
}
