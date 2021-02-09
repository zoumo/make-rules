package golang

import (
	"strings"

	"github.com/zoumo/goset"
)

func (g *GomodHelper) ModTidy() error {
	_, err := g.goRunner.RunCombinedOutput("mod", "tidy")
	if err != nil {
		return err
	}
	return nil
}

// PruneAndTidy
// 1. ensures all existing 'require' directives have an associated 'replace' directive pinning a version
//    - adds explicit 'require' directives for all transitive dependencies
//    - adds explicit 'replace' directives for all require directives
// 2. prune replace directives that pin to the naturally selected version.
// 3. format and tidy go.mod
func (g *GomodHelper) PruneAndTidy() error {
	if err := g.ensureRequireAndReplace(); err != nil {
		g.logger.Error(err, "failed to ensure require and replace")
		return err
	}

	if err := g.ModTidy(); err != nil {
		g.logger.Error(err, "failed to run go mod tidy")
		return err
	}

	if err := g.ensureMissingRequireAndReplace(); err != nil {
		g.logger.Error(err, "failed to ensure missing require and replace")
		return err
	}

	if err := g.Format(); err != nil {
		g.logger.Error(err, "failed to format go.mod")
		return err
	}

	g.logger.Info("====================> go mod tiding <====================")
	if err := g.pruneReplace(); err != nil {
		return err
	}

	if err := g.ModTidy(); err != nil {
		return err
	}

	return nil
}

func (g *GomodHelper) pruneReplace() error {
	// prune replace directives that pin to the naturally selected version.
	// do this before tidying, since tidy removes unused modules that
	// don't provide any relevant packages, which forgets which version of the
	// unused transitive dependency we had a require directive for,
	// and prevents pruning the matching replace directive after tidying.
	listMods, err := g.ParseListMod()
	if err != nil {
		return err
	}

	for _, m := range listMods {
		if m.Replace != nil && m.Path == m.Replace.Path && m.Version == m.Replace.Version &&
			!strings.HasPrefix(m.Path, "k8s.io/") {
			g.logger.Info("drop replace", "reason", "naturally selected", "path", m.Path, "version", m.Version)
			if err := g.EditDropreplace(m.Path); err != nil {
				return err
			}
		}
	}

	// run go mod tidy
	if err := g.ModTidy(); err != nil {
		return err
	}

	// prune unused pinned replace directives
	used := goset.NewSet()
	listMods, err = g.ParseListMod()
	if err != nil {
		return err
	}

	for _, m := range listMods {
		used.Add(m.Path) //nolint:errcheck
	}

	gomod, err := g.ParseMod()
	if err != nil {
		return err
	}

	for _, r := range gomod.Replace {
		if !used.Contains(r.Old.Path) {
			g.logger.Info("drop replace", "reason", "unused", "path", r.Old.Path, "version", r.New.Version)
			// this replace is not found in go list, drop it
			if err := g.EditDropreplace(r.Old.Path); err != nil {
				return err
			}
		}
	}

	// prune replace directives that pin to the naturally selected version
	listMods, err = g.ParseListMod()
	if err != nil {
		return err
	}
	for _, m := range listMods {
		if m.Replace != nil && m.Path == m.Replace.Path && m.Version == m.Replace.Version &&
			!strings.HasPrefix(m.Path, "k8s.io/") {
			g.logger.Info("drop replace", "reason", "naturally selected", "path", m.Path, "version", m.Version)
			if err := g.EditDropreplace(m.Path); err != nil {
				return err
			}
		}
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
func (g *GomodHelper) ensureRequireAndReplace() error {
	// 1. Ensure 'replace' directives have an explicit 'require' directive
	// Filter all 'requires' whose path are not replaced and respect original
	// requires whose path is not replaced to avoid overwrite by replace
	convert := func(req Require, replace Replace) ListModule {
		version := req.Version
		if IsValidVersion(replace.New.Version) &&
			(!IsValidVersion(req.Version) || replace.Old.Path != replace.New.Path) {
			// overwrite version if 'require' a invalid version or path changed
			version = replace.New.Version
		}
		return ListModule{
			Path:     req.Path,
			Version:  version,
			Indirect: req.Indirect,
			Replace: &ListModule{
				Path:    replace.New.Path,
				Version: version,
			},
		}
	}
	mod, err := g.parseGomodToModules(convert)
	if err != nil {
		return err
	}
	if err := g.overwriteGomod(mod); err != nil {
		return err
	}

	// 2. Add explicit require directives for indirect dependencies.
	//    Add explicit replace directives pinning dependencies that aren't pinned yet
	return g.ensureRequireAndReplaceFromGolist()
}

func (g *GomodHelper) ensureMissingRequireAndReplace() error {
	// 1. Add missing require for replace and Add missing replace for require
	convert := func(req Require, replace Replace) ListModule {
		return ListModule{
			Path:     req.Path,
			Version:  req.Version,
			Indirect: req.Indirect,
			Replace: &ListModule{
				Path:    replace.New.Path,
				Version: replace.New.Version,
			},
		}
	}
	mod, err := g.parseGomodToModules(convert)
	if err != nil {
		return err
	}

	if err := g.overwriteGomod(mod); err != nil {
		return err
	}

	// 2. Add explicit require directives for indirect dependencies.
	//    Add explicit replace directives pinning dependencies that aren't pinned yet
	return g.ensureRequireAndReplaceFromGolist()
}

func (g *GomodHelper) parseGomodToModules(conflict func(Require, Replace) ListModule) (map[string]ListModule, error) {
	gomod, err := g.ParseMod()
	if err != nil {
		return nil, err
	}
	// 1. Add missing require for replace and Add missing replace for require
	golist := map[string]ListModule{}
	for _, r := range gomod.Require {
		golist[r.Path] = ListModule{
			Path:     r.Path,
			Version:  r.Version,
			Indirect: r.Indirect,
			Replace: &ListModule{
				Path:    r.Path,
				Version: r.Version,
			},
		}
	}
	for _, r := range gomod.Replace {
		found, ok := golist[r.Old.Path]
		if !ok {
			golist[r.Old.Path] = ListModule{
				Path:    r.Old.Path,
				Version: r.New.Version,
				Replace: &ListModule{
					Path:    r.New.Path,
					Version: r.New.Version,
				},
			}
		} else {
			golist[r.Old.Path] = conflict(Require{Path: found.Path, Version: found.Version, Indirect: found.Indirect}, r)
		}
	}

	return golist, nil
}

func (g *GomodHelper) overwriteGomod(mod map[string]ListModule) error {
	for _, m := range mod {
		if m.Replace == nil {
			// add missing replace
			m.Replace = &ListModule{
				Path:    m.Path,
				Version: m.Version,
			}
		}
		if IsValidVersion(m.Version) {
			if err := g.EditRequire(m.Path, m.Version); err != nil {
				return err
			}
		}
		if IsValidVersion(m.Replace.Version) {
			if err := g.EditReplace(m.Path, m.Replace.Path, m.Replace.Version); err != nil {
				return err
			}
		}
	}
	return nil
}

func (g *GomodHelper) ensureRequireAndReplaceFromGolist() error {
	listMods, err := g.ParseListMod()
	if err != nil {
		return err
	}
	for _, m := range listMods {
		if m.Indirect {
			err := g.EditRequire(m.Path, m.Version)
			if err != nil {
				return err
			}
		}
		if m.Replace == nil {
			err := g.EditReplace(m.Path, m.Path, m.Version)
			if err != nil {
				return err
			}
		}
	}
	return nil
}
