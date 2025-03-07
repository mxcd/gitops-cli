package patch

import (
	"errors"

	"github.com/urfave/cli/v2"
)

type PatchTask struct {
	FilePath string  `json:"filePath"`
	Patches  []Patch `json:"patches"`
}

type Patch struct {
	Selector string `json:"selector"`
	Value    string `json:"value"`
}

type PrepareOptions struct {
	Clone bool
}

type PatchMethod interface {
	Prepare(options *PrepareOptions) error
	Patch(patchTasks []PatchTask) error
}

func PatchCommand(c *cli.Context) error {

	var patchMethod PatchMethod

	if c.String("repository") != "" {
		patcherOptions, err := GetGitPatcherOptionsFromCli(c)
		if err != nil {
			return err
		}

		patcher, err := NewGitPatcher(patcherOptions)
		if err != nil {
			return err
		}

		patchMethod = patcher
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
		FilePath: filePath,
		Patches:  patches,
	}, nil
}
