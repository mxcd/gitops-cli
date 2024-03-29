package util

import (
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetSecretFiles(t *testing.T) {
	secretFiles, err := GetSecretFiles()
	if err != nil {
		t.Error(err)
	}
	if len(secretFiles) == 0 {
		t.Error("No secret files found")
	}
}

func TestGetGitRepoRoot(t *testing.T) {
	rootDir, err := GetGitRepoRoot()
	if err != nil {
		t.Error(err)
	}
	t.Log(rootDir)
}

func TestDecryptFile(t *testing.T) {
	rootDir, _ := GetGitRepoRoot()
	decryptedFile, err := DecryptFile(filepath.Join(rootDir, "test_assets", "test.gitops.secret.enc.yml"))
	if err != nil {
		t.Error(err)
	}
	t.Log(decryptedFile)
}

func TestGetBasenameWithoutExtension(t *testing.T) {
	basename := GetSecretBasename("/foo/bar/baz.gitops.secret.enc.yml")
	assert.Equal(t, "baz", basename, "Basename should be baz")

	basename = GetSecretBasename("/foo/bar/baz.gitops.secret.enc.yaml")
	assert.Equal(t, "baz", basename, "Basename should be baz")

	basename = GetSecretBasename("baz.gitops.secret.enc.yml")
	assert.Equal(t, "baz", basename, "Basename should be baz")
}