package util

import (
	"bufio"
	"fmt"
	"io/fs"
	"math"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"

	log "github.com/sirupsen/logrus"
	"github.com/urfave/cli/v2"
	"go.mozilla.org/sops/v3/decrypt"
)

// go over all files in the current directory (recursively)
// and find all files that end with .secret.enc.yaml or .secret.enc.yml
// return a list of these files
func GetSecretFiles() ([]string, error) {
	log.Trace("Searching for secret files in given directory")

	secretFileRegex, err := regexp.Compile(`.*\.gitops\.secret\.enc\.ya?ml$`)
	if err != nil {
		log.Fatal(err)
	}

	var secretFiles []string
	err = filepath.WalkDir(GetRootDir(),
		func(path string, d fs.DirEntry, err error) error {
			if err != nil {
				return err
			}
			if strings.Contains(path, ".git") && !strings.Contains(path, ".gitops") {
				// log.Trace("Skipping git directory: ", path)
				return nil;
			}

			if d.IsDir() {
				// log.Trace("Skipping directory: ", path)
				return nil;
			}

			if secretFileRegex.MatchString(path) {
				log.Debug("Found secret file: ", path)				
				relativePath, err := filepath.Rel(GetRootDir(), path)
				if err != nil {
					log.Error("An error occurred while getting the relative path of the secret file")
					log.Error(err)
					return err
				}
				relativePath = filepath.ToSlash(relativePath)
				log.Trace("Converted path: ", relativePath)
				secretFiles = append(secretFiles, relativePath)
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

func ToRedactedString(s string ) string {
	return strings.Repeat("*", int(math.Min(float64(len(s)), float64(50))))
}

func StringPrompt(label string) string {
	var s string
	r := bufio.NewReader(os.Stdin)
	for {
			fmt.Fprint(os.Stderr, label+" ")
			s, _ = r.ReadString('\n')
			if s != "" {
					break
			}
	}
	return strings.TrimSpace(s)
}

var cliContext *cli.Context
func SetCliContext(c *cli.Context) {
	cliContext = c
}
func GetCliContext() *cli.Context {
	if cliContext == nil {
		cliContext = GetDummyCliContext()
	}
	return cliContext
}

var _rootDir = ""
func GetRootDir() string {
	if _rootDir == "" {
		ComputeRootDir(GetCliContext())
	}
	return _rootDir
}

func ComputeRootDir(c *cli.Context) {
	if c.String("root-dir") != "" {
		log.Trace("Using root-dir flag")
		_rootDir = c.String("root-dir")
	} else {
		log.Trace("Using git repo root")
		var err error
		_rootDir, err = GetGitRepoRoot()
		if err != nil {
			log.Fatal(err)
		}
		log.Trace("Using root directory: '", _rootDir, "'")

		_, err = os.Stat(_rootDir)
		if os.IsNotExist(err) {
			log.Fatal("Root directory '", _rootDir, "' does not exist")
		}
	}
}

func GetDummyCliContext() *cli.Context {
	app := &cli.App{
		Name:  "gitpos",
		Usage: "GitOps CLI",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:  "root-dir",
				Value: "",
				Usage: "root directory of the git repository",
				EnvVars: []string{"GITOPS_ROOT_DIR"},
			},
		},
	}
	return cli.NewContext(app, nil, nil)
}