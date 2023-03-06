package util

import (
	"fmt"
	"testing"
)

func TestGetSecretFiles(t *testing.T) {
	rootDir, err := GetGitRepoRoot()
	if err != nil {
		t.Error(err)
	}
	secretFiles, err := GetSecretFiles(rootDir)
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
	decryptedFile, err := DecryptFile(fmt.Sprintf("%s%s", rootDir, "/test/test.secret.enc.yml"))
	if err != nil {
		t.Error(err)
	}
	t.Log(decryptedFile)
}