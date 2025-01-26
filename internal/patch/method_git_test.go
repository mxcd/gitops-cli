package patch

import (
	"os"
	"path"
	"testing"

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

func TestGitPath(t *testing.T) {
  sshKey := getSshKeyData(t)

  
}
