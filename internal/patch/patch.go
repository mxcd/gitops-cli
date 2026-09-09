package patch

import (
	"errors"
	"fmt"
	"path/filepath"
	"strings"

	log "github.com/rs/zerolog/log"

	"github.com/urfave/cli/v2"
)

type PatchTask struct {
	Actor    string  `json:"actor"`
	FilePath string  `json:"filePath"`
	Patches  []Patch `json:"patches"`
}

type Patch struct {
	Selector string `json:"selector"`
	Value    string `json:"value"`
}

// FilePatch describes all patches to be applied to a single file.
type FilePatch struct {
	FilePath string  `json:"filePath"`
	Patches  []Patch `json:"patches"`
}

// PatchBatch describes patches for multiple files that are applied and
// committed atomically.
type PatchBatch struct {
	Actor string      `json:"actor"`
	Files []FilePatch `json:"files"`
}

type PrepareOptions struct {
	Clone bool
}

type PatchMethod interface {
	Prepare(options *PrepareOptions) error
	Patch(patchTasks []PatchTask) error
	// PatchBatch applies all patches of the batch in a single commit and
	// returns the commit id. An empty commit id is returned if the batch did
	// not result in any change.
	PatchBatch(batch PatchBatch) (string, error)
}

// ErrInvalidPatchBatch wraps all validation errors of a patch batch so callers
// can distinguish invalid input from execution failures.
var ErrInvalidPatchBatch = errors.New("invalid patch batch")

// cleanRelativeFilePath validates a file path of a patch request and returns
// its cleaned, repository relative form.
func cleanRelativeFilePath(filePath string) (string, error) {
	if strings.TrimSpace(filePath) == "" {
		return "", errors.New("must not be empty")
	}

	if filepath.IsAbs(filePath) || strings.HasPrefix(filePath, "/") || strings.HasPrefix(filePath, "\\") {
		return "", errors.New("must be relative to the repository root")
	}

	// paths are passed to `git add` without a `--` separator, so a leading
	// dash would be interpreted as an option
	if strings.HasPrefix(filePath, "-") {
		return "", errors.New("must not start with '-'")
	}

	cleanedFilePath := filepath.Clean(filePath)
	if cleanedFilePath == "." || cleanedFilePath == ".." || strings.HasPrefix(cleanedFilePath, ".."+string(filepath.Separator)) {
		return "", errors.New("must not escape the repository root")
	}

	return cleanedFilePath, nil
}

// ValidatePatchBatch checks a patch batch for structural errors. All returned
// errors wrap ErrInvalidPatchBatch.
func ValidatePatchBatch(batch PatchBatch) error {
	if len(batch.Files) == 0 {
		return fmt.Errorf("%w: files: must not be empty", ErrInvalidPatchBatch)
	}

	seenFilePaths := map[string]int{}

	for fileIndex, file := range batch.Files {
		cleanedFilePath, err := cleanRelativeFilePath(file.FilePath)
		if err != nil {
			return fmt.Errorf("%w: files[%d].filePath: %s", ErrInvalidPatchBatch, fileIndex, err.Error())
		}

		if previousIndex, ok := seenFilePaths[cleanedFilePath]; ok {
			return fmt.Errorf("%w: files[%d].filePath: duplicate of files[%d].filePath ('%s')", ErrInvalidPatchBatch, fileIndex, previousIndex, cleanedFilePath)
		}
		seenFilePaths[cleanedFilePath] = fileIndex

		if len(file.Patches) == 0 {
			return fmt.Errorf("%w: files[%d].patches: must not be empty", ErrInvalidPatchBatch, fileIndex)
		}

		for patchIndex, filePatch := range file.Patches {
			if strings.TrimSpace(filePatch.Selector) == "" {
				return fmt.Errorf("%w: files[%d].patches[%d].selector: must not be empty", ErrInvalidPatchBatch, fileIndex, patchIndex)
			}
		}
	}

	return nil
}

func PatchCommand(c *cli.Context) error {

	var patchMethod PatchMethod

	if c.String("repository") != "" {
		// patcherOptions, err := GetGitPatcherOptionsFromCli(c)
		// if err != nil {
		// 	return err
		// }

		// patcher, err := NewGitPatcher(patcherOptions)
		// if err != nil {
		// 	return err
		// }

		// patchMethod = patcher
		log.Panic().Msg("Git patcher not implemented")
	} else if c.String("repository-server") != "" {
		method, err := NewRepoServerPatcher(c)
		if err != nil {
			return err
		}
		patchMethod = method
	} else {
		return errors.New("no repository specified")
	}

	err := patchMethod.Prepare(&PrepareOptions{Clone: true})
	if err != nil {
		return err
	}

	patchTask, err := GetPatchTaskFromCli(c)
	if err != nil {
		return err
	}

	err = patchMethod.Patch([]PatchTask{patchTask})
	if err != nil {
		return err
	}

	return nil
}

func GetPatchTaskFromCli(c *cli.Context) (PatchTask, error) {
	filePath := c.Args().First()
	if filePath == "" {
		return PatchTask{}, errors.New("no file specified")
	}

	actor := ""

	if c.String("actor") != "" {
		cliActor := c.String("actor")
		if cliActor != "" {
			actor = cliActor
		}
	}

	patches := []Patch{}
	args := c.Args().Tail()

	if len(args)%2 != 0 {
		return PatchTask{}, errors.New("invalid number of arguments. patches must be in the form of 'selector value'")
	}

	for i := 0; i < len(args); i += 2 {
		patches = append(patches, Patch{
			Selector: args[i],
			Value:    args[i+1],
		})
	}

	return PatchTask{
		Actor:    actor,
		FilePath: filePath,
		Patches:  patches,
	}, nil
}
