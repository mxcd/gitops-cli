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

func TestGitSshPatch(t *testing.T) {
	sshKey := getSshKeyData(t)

	options := &GitPatcherOptions{
		Branch:                  "main",
		RepositoryUrl:           "ssh://localhost:23231/gitops-test.git",
		SshPrivateKey:           sshKey,
		NoStrictHostKeyChecking: true,
	}

	patcher, err := NewGitPatcher(options)
	assert.NoError(t, err)

	err = patcher.Prepare()
	assert.NoError(t, err)

	patchTask := PatchTask{
		FilePath: "applications/dev/service-test/values.yaml",
		Patches: []Patch{
			{
				Selector: ".service.image.tag",
				Value:    "v1.0.0",
			},
		},
	}

	err = patcher.Patch([]PatchTask{patchTask})
	assert.NoError(t, err)
}
