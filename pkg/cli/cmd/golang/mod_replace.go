package golang

import (
	"fmt"
	"path"

	"github.com/spf13/pflag"

	"github.com/zoumo/make-rules/pkg/cli/injection"
	"github.com/zoumo/make-rules/pkg/cli/plugin"
	"github.com/zoumo/make-rules/pkg/golang"
)

type modReplaceSubcommand struct {
	*injection.InjectionMixin
	gomod *golang.GomodHelper
}

func NewModReplaceCommand() plugin.Subcommand {
	return &modReplaceSubcommand{
		InjectionMixin: injection.NewInjectionMixin(),
	}
}

func (c *modReplaceSubcommand) Name() string {
	return "replace"
}

func (c *modReplaceSubcommand) BindFlags(fs *pflag.FlagSet) {
}

func (c *modReplaceSubcommand) PreRun(args []string) error {
	if len(args) != 2 && len(args) != 3 {
		return fmt.Errorf("require module path and version")
	}
	modfile := path.Join(c.Workspace, "go.mod")
	c.gomod = golang.NewGomodHelper(modfile, c.Logger)

	return nil
}

func (c *modReplaceSubcommand) Run(args []string) error {
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
