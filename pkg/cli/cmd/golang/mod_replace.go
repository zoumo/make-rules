package golang

import (
	"fmt"
	"path"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/zoumo/golib/cli"

	"github.com/zoumo/make-rules/pkg/cli/common"
	"github.com/zoumo/make-rules/pkg/golang"
)

var _ cli.Command = &ModReplaceCommand{}
var _ cli.ComplexOptions = &ModReplaceCommand{}

type ModReplaceCommand struct {
	*common.CommonOptions
	gomod *golang.GomodHelper
}

func NewModReplaceCommand() *cobra.Command {
	return cli.NewCobraCommand(&ModReplaceCommand{
		CommonOptions: common.NewCommonOptions(),
	})
}

func (c *ModReplaceCommand) Name() string {
	return "replace"
}

func (c *ModReplaceCommand) BindFlags(fs *pflag.FlagSet) {
	c.CommonOptions.BindFlags(fs)
}

func (c *ModReplaceCommand) Complete(cmd *cobra.Command, args []string) error {
	if err := c.CommonOptions.Complete(cmd, args); err != nil {
		return err
	}
	modfile := path.Join(c.Workspace, "go.mod")
	c.gomod = golang.NewGomodHelper(modfile, c.Logger)
	return nil
}

func (c *ModReplaceCommand) Validate() error {
	return c.CommonOptions.Validate()
}

func (c *ModReplaceCommand) Run(cmd *cobra.Command, args []string) error {
	if len(args) != 2 && len(args) != 3 {
		return fmt.Errorf("require module path and version")
	}
	var path, newPath, version string
	if len(args) == 2 {
		path = args[0]
		newPath = path
		version = args[1]
	} else {
		path = args[0]
		newPath = args[1]
		version = args[2]
	}
	return c.gomod.Replace(path, newPath, version)
}
