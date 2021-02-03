package golang

import (
	"path"

	"github.com/spf13/pflag"

	"github.com/zoumo/make-rules/pkg/cli/injection"
	"github.com/zoumo/make-rules/pkg/cli/plugin"
	"github.com/zoumo/make-rules/pkg/golang"
)

type tidySubcommand struct {
	*injection.InjectionMixin
	gomod *golang.GomodHelper
}

func NewModTidyCommand() plugin.Subcommand {
	return &tidySubcommand{
		InjectionMixin: &injection.InjectionMixin{},
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
	if err := c.ensureRequireAndReplace(); err != nil {
		c.Logger.Error(err, "failed to ensure require and replace")
		return err
	}

	if err := c.gomod.ModTidy(); err != nil {
		c.Logger.Error(err, "failed to run go mod tidy")
		return err
	}

	if err := c.ensureMissingRequireAndReplace(); err != nil {
		c.Logger.Error(err, "failed to ensure missing require and replace")
		return err
	}

	if err := c.gomod.Format(); err != nil {
		c.Logger.Error(err, "failed to format go.mod")
		return err
	}

	c.Logger.Info(">> go mod tiding")
	if err := c.gomod.PruneReplace(); err != nil {
		return err
	}

	if err := c.gomod.ModTidy(); err != nil {
		return err
	}

	return nil
}

// - ensures all existing 'require' directives have an associated 'replace' directive pinning a version
// - adds explicit 'require' directives for all transitive dependencies
// - adds explicit 'replace' directives for all require directives
//
// It picks version follows these precedence:
// 1. 'replace' with path changed, e.g. require github.com/a/foo v2.0.0 but replace github.com/a/foo => github.com/b/foo v1.0.0 use v1.0.0
// 2. 'require' with valid version, e.g. require k8s.io/api v0.18.0 but replace k8s.io/api => k8s.io/api v0.17.0, use v0.18.0
// 3. 'replace' with valid version, e.g. require k8s.io/api v0.0.0 and replace k8s.io/api => k8s.io/api v0.18.0, use v0.18.0
func (c *tidySubcommand) ensureRequireAndReplace() error {
	gomod, err := c.gomod.ParseMod()
	if err != nil {
		return err
	}
	// 1. Ensure 'replace' directives have an explicit 'require' directive
	// Filter all 'requires' whose path are not replaced and respect original
	// requires whose path is not replaced to avoid overwrite by replace
	golist := map[string]golang.ListModule{}
	for _, r := range gomod.Require {
		golist[r.Path] = golang.ListModule{
			Path:     r.Path,
			Version:  r.Version,
			Indirect: r.Indirect,
		}
	}
	for _, r := range gomod.Replace {
		found, ok := golist[r.Old.Path]
		if ok {
			if !golang.IsValidVersion(found.Version) ||
				(r.Old.Path != r.New.Path && golang.IsValidVersion(r.New.Version)) {
				// overwrite version: invalid version or path changed
				found.Version = r.New.Version
			}
			found.Replace = &golang.ListModule{
				Path:    r.New.Path,
				Version: r.New.Version,
			}
			golist[r.Old.Path] = found
		} else {
			golist[r.Old.Path] = golang.ListModule{
				Path:    r.Old.Path,
				Version: r.New.Version,
				Replace: &golang.ListModule{
					Path:    r.New.Path,
					Version: r.New.Version,
				},
			}
		}
	}

	for _, m := range golist {
		if m.Replace == nil {
			// add missing replace
			m.Replace = &golang.ListModule{
				Path:    m.Path,
				Version: m.Version,
			}
		}
		if golang.IsValidVersion(m.Version) {
			if err := c.gomod.EditRequire(m.Path, m.Version); err != nil {
				return err
			}
		}
		if golang.IsValidVersion(m.Replace.Version) {
			if err := c.gomod.EditReplace(m.Path, m.Replace.Path, m.Replace.Version); err != nil {
				return err
			}
		}
	}

	// 2. Add explicit require directives for indirect dependencies.
	//    Add explicit replace directives pinning dependencies that aren't pinned yet
	return c.ensureRequireAndReplaceFromGolist()
}

func (c *tidySubcommand) ensureMissingRequireAndReplace() error {
	gomod, err := c.gomod.ParseMod()
	if err != nil {
		return err
	}
	// 1. Add missing require for replace and Add missing replace for require
	golist := map[string]golang.ListModule{}
	for _, r := range gomod.Require {
		golist[r.Path] = golang.ListModule{
			Path:     r.Path,
			Version:  r.Version,
			Indirect: r.Indirect,
		}
	}
	for _, r := range gomod.Replace {
		_, ok := golist[r.Old.Path]
		if !ok {
			golist[r.Old.Path] = golang.ListModule{
				Path:    r.Old.Path,
				Version: r.Old.Version,
				Replace: &golang.ListModule{
					Path:    r.New.Path,
					Version: r.New.Version,
				},
			}
		}
	}

	for _, m := range golist {
		if m.Replace == nil {
			// add missing replace
			m.Replace = &golang.ListModule{
				Path:    m.Path,
				Version: m.Version,
			}
		}
		if golang.IsValidVersion(m.Version) {
			if err := c.gomod.EditRequire(m.Path, m.Version); err != nil {
				return err
			}
		}
		if m.Replace.Version != "" {
			if err := c.gomod.EditReplace(m.Path, m.Replace.Path, m.Replace.Version); err != nil {
				return err
			}
		}
	}

	// 2. Add explicit require directives for indirect dependencies.
	//    Add explicit replace directives pinning dependencies that aren't pinned yet
	return c.ensureRequireAndReplaceFromGolist()
}

func (c *tidySubcommand) ensureRequireAndReplaceFromGolist() error {
	listMods, err := c.gomod.ParseListMod()
	if err != nil {
		c.Logger.Error(err, "failed to get go list deps")
		return err
	}
	for _, m := range listMods {
		if m.Indirect {
			err := c.gomod.EditRequire(m.Path, m.Version)
			if err != nil {
				return err
			}
		}
		if m.Replace == nil {
			err := c.gomod.EditReplace(m.Path, m.Path, m.Version)
			if err != nil {
				return err
			}
		}
	}
	return nil
}
