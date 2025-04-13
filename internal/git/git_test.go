package git

// import (
// 	"log"
// 	"os"
// 	"path"
// 	"testing"

// 	"github.com/google/uuid"
// 	"github.com/mxcd/gitops-cli/internal/util"
// 	"github.com/stretchr/testify/assert"
// )

// func getSshKeyData(t *testing.T) []byte {
// 	baseDir, err := util.GetGitRepoRoot()
// 	assert.NoError(t, err)
// 	assert.NotEmpty(t, baseDir)

// 	sshKeyPath := path.Join(baseDir, "hack", "soft-serve", "ssh-key")
// 	assert.FileExists(t, sshKeyPath)

// 	sshKey, err := os.ReadFile(sshKeyPath)
// 	assert.NoError(t, err)
// 	assert.NotEmpty(t, sshKey)

// 	return sshKey
// }

// func TestNewGitConnection(t *testing.T) {

// 	sshKey := getSshKeyData(t)

// 	baseDir, err := util.GetGitRepoRoot()
// 	assert.NoError(t, err)

// 	authentication, err := GetAuthFromSshKey(sshKey, nil)
// 	assert.NoError(t, err)
// 	assert.NotNil(t, authentication)

// 	uuid := uuid.New().String()

// 	options := &ConnectionOptions{
// 		Directory:      path.Join(baseDir, "sandbox", "gitops-test-"+uuid),
// 		Repository:     "ssh://git@localhost:23231/gitops-test.git",
// 		Branch:         "main",
// 		Authentication: authentication,
// 	}
// 	gitConnection, err := NewGitConnection(options)
// 	assert.NoError(t, err)
// 	assert.NotNil(t, gitConnection)

// 	err = gitConnection.Clone()
// 	assert.NoError(t, err)
// }

// func cloneTempRepository(t *testing.T) *Connection {
// 	sshKey := getSshKeyData(t)

// 	baseDir, err := util.GetGitRepoRoot()
// 	assert.NoError(t, err)

// 	uuid := uuid.New().String()
// 	directoryName := path.Join(baseDir, "sandbox", "gitops-test-"+uuid)

// 	authentication, err := GetAuthFromSshKey(sshKey, nil)
// 	assert.NoError(t, err)
// 	assert.NotNil(t, authentication)

// 	options := &ConnectionOptions{
// 		Directory:      directoryName,
// 		Repository:     "ssh://git@localhost:23231/gitops-test.git",
// 		Branch:         "main",
// 		Authentication: authentication,
// 	}
// 	gitConnection, err := NewGitConnection(options)
// 	assert.NoError(t, err)
// 	assert.NotNil(t, gitConnection)

// 	err = gitConnection.Clone()
// 	assert.NoError(t, err)

// 	return gitConnection
// }

// func TestGitPullFastForward(t *testing.T) {

// 	tempConnectionA := cloneTempRepository(t)
// 	assert.NotNil(t, tempConnectionA)
// 	log.Println("tempConnectionA cloned")

// 	tempConnectionB := cloneTempRepository(t)
// 	assert.NotNil(t, tempConnectionB)
// 	log.Println("tempConnectionB cloned")

// 	uuid := uuid.New().String()
// 	testFileName := "test-file-" + uuid
// 	testFilePath := path.Join(tempConnectionA.Options.Directory, testFileName)
// 	err := os.WriteFile(testFilePath, []byte("test"), 0644)
// 	assert.NoError(t, err)

// 	log.Printf("file written to %s", testFilePath)

// 	hash, err := tempConnectionA.Commit([]string{testFileName}, "Test commit")
// 	assert.NoError(t, err)
// 	assert.NotEmpty(t, hash)

// 	err = tempConnectionA.Push()
// 	assert.NoError(t, err)

// 	err = tempConnectionB.Pull()
// 	assert.NoError(t, err)

// 	testFilePath = path.Join(tempConnectionB.Options.Directory, testFileName)
// 	_, err = os.Stat(testFilePath)
// 	assert.NoError(t, err)
// 	data, err := os.ReadFile(testFilePath)
// 	assert.NoError(t, err)
// 	assert.Equal(t, "test", string(data))
// }

// func TestGitPullRebase(t *testing.T) {

// 	// Clone repository A
// 	tempConnectionA := cloneTempRepository(t)
// 	assert.NotNil(t, tempConnectionA)
// 	log.Println("tempConnectionA cloned")

// 	// Clone repository B
// 	tempConnectionB := cloneTempRepository(t)
// 	assert.NotNil(t, tempConnectionB)
// 	log.Println("tempConnectionB cloned")

// 	// Create a new file in repository A, commit and push it
// 	uuidA := uuid.New().String()
// 	testFileNameA := "test-file-" + uuidA
// 	testFilePathA := path.Join(tempConnectionA.Options.Directory, testFileNameA)
// 	err := os.WriteFile(testFilePathA, []byte("test A"), 0644)
// 	assert.NoError(t, err)

// 	log.Printf("file written to %s", testFilePathA)

// 	hashA, err := tempConnectionA.Commit([]string{testFileNameA}, "Test commit A")
// 	assert.NoError(t, err)
// 	assert.NotEmpty(t, hashA)

// 	err = tempConnectionA.Push()
// 	assert.NoError(t, err)

// 	// Create a new file in repository B, commit and push it
// 	uuidB := uuid.New().String()
// 	testFileNameB := "test-file-" + uuidB
// 	testFilePathB := path.Join(tempConnectionB.Options.Directory, testFileNameB)
// 	err = os.WriteFile(testFilePathB, []byte("test B"), 0644)
// 	assert.NoError(t, err)

// 	log.Printf("file written to %s", testFilePathB)

// 	hashB, err := tempConnectionB.Commit([]string{testFileNameB}, "Test commit B")
// 	assert.NoError(t, err)
// 	assert.NotEmpty(t, hashB)

// 	// Push is expected to fail because of changes from repository A in remote
// 	err = tempConnectionB.Push()
// 	assert.Error(t, err)

// 	// Pull is expected to rebase the changes from repository A
// 	err = tempConnectionB.Pull()
// 	assert.NoError(t, err)

// 	// check if test file A is added to repo B
// 	testFilePath := path.Join(tempConnectionB.Options.Directory, testFileNameA)
// 	_, err = os.Stat(testFilePath)
// 	assert.NoError(t, err)
// 	data, err := os.ReadFile(testFilePath)
// 	assert.NoError(t, err)
// 	assert.Equal(t, "test A", string(data))

// 	// Push is expected to succeed after rebase
// 	err = tempConnectionB.Push()
// 	assert.NoError(t, err)

// 	// Pull is expected to pull the changes from repository B
// 	err = tempConnectionA.Pull()
// 	assert.NoError(t, err)

// 	// check if test file B is added to repo A
// 	testFilePath = path.Join(tempConnectionA.Options.Directory, testFileNameB)
// 	_, err = os.Stat(testFilePath)
// 	assert.NoError(t, err)
// 	data, err = os.ReadFile(testFilePath)
// 	assert.NoError(t, err)
// 	assert.Equal(t, "test B", string(data))
// }
