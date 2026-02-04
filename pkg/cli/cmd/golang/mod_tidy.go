package golang

import (
	"path"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/zoumo/golib/cli"

	"github.com/zoumo/make-rules/pkg/cli/common"
	"github.com/zoumo/make-rules/pkg/golang"
)

var _ cli.Command = &ModTidyCommand{}
var _ cli.ComplexOptions = &ModTidyCommand{}

type ModTidyCommand struct {
	*common.CommonOptions
	gomod *golang.GomodHelper
}

func NewModTidyCommand() *cobra.Command {
	return cli.NewCobraCommand(&ModTidyCommand{
		CommonOptions: common.NewCommonOptions(),
	})
}

func (c *ModTidyCommand) Name() string {
	return "tidy"
}

func (c *ModTidyCommand) BindFlags(fs *pflag.FlagSet) {
	c.CommonOptions.BindFlags(fs)
}

func (c *ModTidyCommand) Complete(cmd *cobra.Command, args []string) error {
	if err := c.CommonOptions.Complete(cmd, args); err != nil {
		return err
	}
	modfile := path.Join(c.Workspace, "go.mod")
	c.gomod = golang.NewGomodHelper(modfile, c.Logger)
	return nil
}

func (c *ModTidyCommand) Validate() error {
	return c.CommonOptions.Validate()
}

func (c *ModTidyCommand) Run(cmd *cobra.Command, args []string) error {
	return c.gomod.PruneAndTidy()
}
