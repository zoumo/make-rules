package golang

import (
	"fmt"
	"path"

	"github.com/spf13/pflag"

	"github.com/zoumo/make-rules/pkg/cli/injection"
	"github.com/zoumo/make-rules/pkg/cli/plugin"
	"github.com/zoumo/make-rules/pkg/golang"
)

type modRequireSubcommand struct {
	*injection.InjectionMixin
	gomod    *golang.GomodHelper
	skipDeps bool
}

func NewModRequireCommand() plugin.Subcommand {
	return &modRequireSubcommand{
		InjectionMixin: injection.NewInjectionMixin(),
	}
}

func (c *modRequireSubcommand) Name() string {
	return "require"
}

func (c *modRequireSubcommand) BindFlags(fs *pflag.FlagSet) {
	fs.BoolVar(&c.skipDeps, "skip-deps", c.skipDeps, "skip require dependencies of input modules")
}

func (c *modRequireSubcommand) PreRun(args []string) error {
	if len(args) != 2 {
		return fmt.Errorf("required module and version")
	}
	modfile := path.Join(c.Workspace, "go.mod")
	c.gomod = golang.NewGomodHelper(modfile, c.Logger)

	return nil
}

func (c *modRequireSubcommand) Run(args []string) error {
	path := args[0]
	version := args[1]
	return c.gomod.Require(path, version, c.skipDeps)
}