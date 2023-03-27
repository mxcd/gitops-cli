package util

import (
	"io/fs"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"

	log "github.com/sirupsen/logrus"
	"go.mozilla.org/sops/v3/decrypt"
)

// go over all files in the current directory (recursively)
// and find all files that end with .secret.enc.yaml or .secret.enc.yml
// return a list of these files
func GetSecretFiles(rootDirectory string) ([]string, error) {
	log.Trace("Searching for secret files in given directory")

	secretFileRegex, err := regexp.Compile(`.*\.gitops\.secret\.enc\.ya?ml$`)
	if err != nil {
		log.Fatal(err)
	}

	var secretFiles []string
	err = filepath.WalkDir(rootDirectory,
		func(path string, d fs.DirEntry, err error) error {
			if err != nil {
				return err
			}
			if strings.Contains(path, ".git") && !strings.Contains(path, ".gitops") {
				log.Trace("Skipping git directory: ", path)
				return nil;
			}

			if d.IsDir() {
				log.Trace("Skipping directory: ", path)
				return nil;
			}

			if secretFileRegex.MatchString(path) {
				log.Debug("Found secret file: ", path)
				secretFiles = append(secretFiles, path)
			}
			return nil
		})
	if err != nil {
		log.Error("An error occurred while searching for secret files")
		log.Error(err)
		return nil, err
	}
	return secretFiles, nil
}

func GetGitRepoRoot() (string, error) {
	log.Trace("Searching for git repo root")
	path, err := exec.Command("git", "rev-parse", "--show-toplevel").Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(path)), nil
}

func DecryptFile(path string) ([]byte, error) {
	log.Trace("Decrypting file: ", path)
	decrypted, err := decrypt.File(path, "yaml")
	if err != nil {
		return []byte{}, err
	}
	return decrypted, nil
}


var secretFilenameRegex = regexp.MustCompile(`\.gitops\.secret\.enc\.ya?ml$`)

/* 
Removes the path and the `gitops.secret.enc.ya?ml` suffix from a given path
*/
func GetSecretBasename(path string) string {
	return secretFilenameRegex.ReplaceAllString(filepath.Base(path), "")
}