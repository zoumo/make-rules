package golang

import (
	"path"

	"github.com/spf13/pflag"
	"github.com/zoumo/golib/cli/plugin"

	"github.com/zoumo/make-rules/pkg/cli/injection"
	"github.com/zoumo/make-rules/pkg/golang"
)

type modUpdateSubcommand struct {
	*injection.InjectionMixin
	gomod *golang.GomodHelper
}

func NewModUpdateCommand() plugin.Subcommand {
	return &modUpdateSubcommand{
		InjectionMixin: injection.NewInjectionMixin(),
	}
}

func (c *modUpdateSubcommand) Name() string {
	return "update"
}

func (c *modUpdateSubcommand) BindFlags(fs *pflag.FlagSet) {
}

func (c *modUpdateSubcommand) PreRun(args []string) error {
	modfile := path.Join(c.Workspace, "go.mod")
	c.gomod = golang.NewGomodHelper(modfile, c.Logger)
	return nil
}

func (c *modUpdateSubcommand) Run(args []string) error {
	// ensure requires
	for _, r := range c.Config.Go.Mod.Require {
		if err := c.gomod.Require(r.Path, r.Version, false); err != nil {
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
