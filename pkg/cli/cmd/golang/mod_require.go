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

var _ cli.Command = &ModRequireCommand{}
var _ cli.ComplexOptions = &ModRequireCommand{}

type ModRequireCommand struct {
	*common.CommonOptions
	gomod *golang.GomodHelper
}

func NewModRequireCommand() *cobra.Command {
	return cli.NewCobraCommand(&ModRequireCommand{
		CommonOptions: common.NewCommonOptions(),
	})
}

func (c *ModRequireCommand) Name() string {
	return "require"
}

func (c *ModRequireCommand) BindFlags(fs *pflag.FlagSet) {
	c.CommonOptions.BindFlags(fs)
}

func (c *ModRequireCommand) Complete(cmd *cobra.Command, args []string) error {
	if err := c.CommonOptions.Complete(cmd, args); err != nil {
		return err
	}
	modfile := path.Join(c.Workspace, "go.mod")
	c.gomod = golang.NewGomodHelper(modfile, c.Logger)
	return nil
}

func (c *ModRequireCommand) Validate() error {
	return c.CommonOptions.Validate()
}

func (c *ModRequireCommand) Run(cmd *cobra.Command, args []string) error {
	if len(args) != 2 {
		return fmt.Errorf("required module and version")
	}
	path := args[0]
	version := args[1]
	return c.gomod.Require(path, version, false)
}
