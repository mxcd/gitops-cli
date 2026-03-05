package secret

import (
	"errors"

	"github.com/urfave/cli/v2"
)

func CompareCommand(c *cli.Context) error {
	if c.Args().Len() != 2 {
		return errors.New("usage: gitops secrets compare <secret-path-1> <secret-path-2>")
	}

	firstPath := c.Args().Get(0)
	secondPath := c.Args().Get(1)

	firstSecret, err := FromPath(firstPath)
	if err != nil {
		return err
	}

	secondSecret, err := FromPath(secondPath)
	if err != nil {
		return err
	}

	diff := CompareSecrets(firstSecret, secondSecret)
	diff.Print(c.Bool("show-unchanged"))

	return nil
}
