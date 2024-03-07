package finalizer

import (
	"github.com/mxcd/gitops-cli/internal/state"
	"github.com/urfave/cli/v2"
)

func ExitApplication(c *cli.Context, writeState bool) error {
	var err error
	if writeState {
		err = state.GetState().Save(c)
	}
	return err
}
