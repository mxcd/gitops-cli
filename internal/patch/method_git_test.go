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

func TestGitSshPatch(t *testing.T) {
	sshKey := getSshKeyData(t)
	baseDir, err := util.GetGitRepoRoot()
	assert.NoError(t, err)

	uuidA := uuid.New().String()
	repositoryPathA := path.Join(baseDir, "sandbox", "gitops-test-"+uuidA)
	err = os.MkdirAll(repositoryPathA, 0755)
	assert.NoError(t, err)

	gitConnectionOptionsA := &git.ConnectionOptions{
		Repository:       "ssh://localhost:23231/gitops-test.git",
		Directory:        repositoryPathA,
		Branch:           "main",
		IgnoreSshHostKey: true,
		Authentication: &git.Authentication{
			SshKey: &git.SshKey{
				PrivateKey: sshKey,
			},
		},
	}

	patcher, err := NewGitPatcher(&GitPatcherOptions{
		GitConnectionOptions: gitConnectionOptionsA,
	})
	assert.NoError(t, err)
	assert.NotNil(t, patcher)

	err = patcher.Prepare(&PrepareOptions{
		Clone: true,
	})
	assert.NotNil(t, patcher.GitConnection)
	assert.NoError(t, err)

	patchTask := PatchTask{
		FilePath: "applications/dev/service-test/values.yaml",
		Patches: []Patch{
			{
				Selector: ".service.image.tag",
				Value:    "v1.0.1",
			},
		},
	}

	err = patcher.Patch([]PatchTask{patchTask})
	assert.NoError(t, err)
}
