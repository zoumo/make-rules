package golang

import (
	"path"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/zoumo/golib/cli"

	"github.com/zoumo/make-rules/pkg/cli/common"
	"github.com/zoumo/make-rules/pkg/golang"
)

var _ cli.Command = &ModUpdateCommand{}
var _ cli.ComplexOptions = &ModUpdateCommand{}

type ModUpdateCommand struct {
	*common.CommonOptions
	gomod *golang.GomodHelper
}

func NewModUpdateCommand() *cobra.Command {
	return cli.NewCobraCommand(&ModUpdateCommand{
		CommonOptions: common.NewCommonOptions(),
	})
}

func (c *ModUpdateCommand) Name() string {
	return "update"
}

func (c *ModUpdateCommand) BindFlags(fs *pflag.FlagSet) {
	c.CommonOptions.BindFlags(fs)
}

func (c *ModUpdateCommand) Complete(cmd *cobra.Command, args []string) error {
	if err := c.CommonOptions.Complete(cmd, args); err != nil {
		return err
	}
	modfile := path.Join(c.Workspace, "go.mod")
	c.gomod = golang.NewGomodHelper(modfile, c.Logger)
	return nil
}

func (c *ModUpdateCommand) Validate() error {
	return c.CommonOptions.Validate()
}

func (c *ModUpdateCommand) Run(cmd *cobra.Command, args []string) error {
	// ensure requires
	for _, r := range c.Config.Go.Mod.Require {
		if err := c.gomod.Require(r.Path, r.Version, r.SkipDeps); err != nil {
			return err
		}
	}
	// ensure replace
	for _, r := range c.Config.Go.Mod.Replace {
		newPath := r.OldPath
		if r.NewPath != "" {
			newPath = r.NewPath
		}
		if err := c.gomod.Replace(r.OldPath, newPath, r.Version); err != nil {
			return err
		}
	}
	// prune and tidy
	return c.gomod.PruneAndTidy()
}
