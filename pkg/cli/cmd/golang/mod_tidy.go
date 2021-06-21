package golang

import (
	"path"

	"github.com/spf13/pflag"
	"github.com/zoumo/golib/cli/plugin"

	"github.com/zoumo/make-rules/pkg/cli/injection"
	"github.com/zoumo/make-rules/pkg/golang"
)

type tidySubcommand struct {
	*injection.InjectionMixin
	gomod *golang.GomodHelper
}

func NewModTidyCommand() plugin.Subcommand {
	return &tidySubcommand{
		InjectionMixin: injection.NewInjectionMixin(),
	}
}

func (c *tidySubcommand) Name() string {
	return "tidy"
}

func (c *tidySubcommand) BindFlags(fs *pflag.FlagSet) {
}

func (c *tidySubcommand) PreRun(args []string) error {
	modfile := path.Join(c.Workspace, "go.mod")
	c.gomod = golang.NewGomodHelper(modfile, c.Logger)

	return nil
}

func (c *tidySubcommand) Run(args []string) error {
	return c.gomod.PruneAndTidy()
}
