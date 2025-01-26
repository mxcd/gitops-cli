package patch

import (
	"errors"

	"github.com/urfave/cli/v2"
)

type PatchTask struct {
	FilePath string
	Patches  []Patch
}

type Patch struct {
	Selector string
	Value    string
}

type PatchMethod interface {
	Prepare() error
	Patch(patchTasks []PatchTask) error
}

func PatchCommand(c *cli.Context) error {

	var patchMethod PatchMethod

	if c.String("repo") != "" {
		patcherOptions, err := GetGitPatcherOptionsFromCli(c)
		if err != nil {
			return err
		}

		patcher, err := NewGitPatcher(patcherOptions)
		if err != nil {
			return err
		}

		patchMethod = patcher
	} else if c.String("repo-server") != "" {
		// return PatchMethodRepoServer(c)
	} else {
		return errors.New("no repository specified")
	}

	err := patchMethod.Prepare()
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
