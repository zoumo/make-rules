package golang

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"path"
	"strings"

	"github.com/go-logr/logr"
	"github.com/zoumo/goset"

	"github.com/zoumo/make-rules/pkg/runner"
)

type ModeEditVerb string

const (
	EditRequire     ModeEditVerb = "require"
	EditReplace     ModeEditVerb = "replace"
	EditDropreplace ModeEditVerb = "dropreplace"
)

type GomodHelper struct {
	goRunner     *runner.Runner
	modfile      string
	logger       logr.Logger
	downloadTemp string
}

func NewGomodHelper(modfile string, logger logr.Logger) *GomodHelper {
	g := &GomodHelper{
		modfile:  modfile,
		goRunner: runner.NewRunner("go").WithDir(path.Dir(modfile)),
		logger:   logger,
	}
	return g
}

func (g *GomodHelper) Require(path, version string) error {
	mod, err := g.ModDownload(path, version)
	if err != nil {
		return err
	}

	// copy modfile to temp
	modfile := backupGomod(mod.GoMod)
	if modfile == "" {
		return fmt.Errorf("module %s version %s does not have a go.mod file", path, version)
	}

	g.logger.Info("download go.mod", "filepath", modfile)

	newgomod := NewGomodHelper(modfile, g.logger)
	if path == "k8s.io/kubernetes" {
		gomod, err := newgomod.ParseMod()
		if err != nil {
			return err
		}
		if err := g.replaceK8SStaging(gomod, version); err != nil {
			return err
		}
	}
	// find local list deps
	requireMod, err := newgomod.ParseMod()
	if err != nil {
		return err
	}

	for _, r := range requireMod.Require {
		if IsValidVersion(r.Version) {
			g.logger.Info(fmt.Sprintf("require %s@%s", r.Path, r.Version), "source", "require")
			if err := g.EditRequire(r.Path, r.Version); err != nil {
				return err
			}
		}
	}

	for _, r := range requireMod.Replace {
		if IsValidVersion(r.New.Version) {
			g.logger.Info(fmt.Sprintf("require %s@%s", r.Old.Path, r.New.Version), "source", "replace")
			if err := g.EditRequire(r.Old.Path, r.New.Version); err != nil {
				return err
			}
		}
	}

	if err := g.EditRequire(path, version); err != nil {
		return err
	}
	if err := g.EditReplace(path, path, version); err != nil {
		return err
	}
	return nil
}

func (g *GomodHelper) replaceK8SStaging(gomod *GoMod, version string) error {
	version = strings.TrimLeft(version, "v")

	staging := []string{}
	for _, r := range gomod.Replace {
		if strings.HasPrefix(r.New.Path, "./staging/src/k8s.io/") {
			staging = append(staging, r.Old.Path)
		}
	}

	// replace staging module
	for _, path := range staging {
		// find staging version
		mod, err := g.ModDownload(path, "kubernetes-"+version)
		if err != nil {
			return err
		}
		g.logger.Info(fmt.Sprintf("replace k8s staging %s=%s@%s", path, path, mod.Version))
		if err := g.EditRequire(path, mod.Version); err != nil {
			return err
		}
		if err := g.EditReplace(path, path, mod.Version); err != nil {
			return err
		}
	}
	return nil
}

func (g *GomodHelper) Format() error {
	gomod, err := g.ParseMod()
	if err != nil {
		return err
	}

	buf := &bytes.Buffer{}

	buf.WriteString(fmt.Sprintf("module %s\n", gomod.Module.Path))
	buf.WriteString("\n")
	buf.WriteString(fmt.Sprintf("go %s", gomod.Go))
	buf.WriteString("\n")
	if len(gomod.Require) > 0 {
		buf.WriteString("require (\n")
		for _, r := range gomod.Require {
			buf.WriteString(fmt.Sprintf("    %s %s", r.Path, r.Version))
			if r.Indirect {
				buf.WriteString(" // indirect")
			}
			buf.WriteString("\n")
		}
		buf.WriteString(")\n")
		buf.WriteString("\n")
	}
	if len(gomod.Exclude) > 0 {
		buf.WriteString("exclude (\n")
		for _, r := range gomod.Exclude {
			buf.WriteString(fmt.Sprintf("    %s %s\n", r.Path, r.Version))
		}
		buf.WriteString(")\n")
		buf.WriteString("\n")
	}
	if len(gomod.Replace) > 0 {
		buf.WriteString("replace (\n")
		for _, r := range gomod.Replace {
			buf.WriteString(fmt.Sprintf("    %s => %s %s\n", r.Old.Path, r.New.Path, r.New.Version))
		}
		buf.WriteString(")\n")
	}

	// overwrite go.mod
	return ioutil.WriteFile(g.modfile, buf.Bytes(), 0644)
}

func (g *GomodHelper) ParseMod() (*GoMod, error) {
	data, err := g.goRunner.RunOutput("mod", "edit", "-json")
	if err != nil {
		return nil, err
	}

	gomod := GoMod{}
	if err := json.Unmarshal(data, &gomod); err != nil {
		return nil, err
	}

	return &gomod, nil
}

func (g *GomodHelper) ParseListMod() ([]ListModule, error) {
	// backup go.mod because go list will change go.mod
	file := backupGomod(g.modfile)

	out, err := g.goRunner.RunOutput("list", "-m", "-json", "all")
	if err != nil {
		return nil, err
	}

	decoder := json.NewDecoder(bytes.NewBuffer(out))
	ret := []ListModule{}
	for {
		var m ListModule
		if err := decoder.Decode(&m); err == io.EOF {
			break
		} else if err != nil {
			return nil, err
		}
		if m.Main {
			// skip main
			continue
		}
		ret = append(ret, m)
	}

	restoreGomod(g.modfile, file)
	return ret, nil
}

func (g *GomodHelper) ModTidy() error {
	_, err := g.goRunner.RunCombinedOutput("mod", "tidy")
	if err != nil {
		return err
	}
	return nil
}

func (g *GomodHelper) PruneReplace() error {
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
			g.logger.Info("drop replace", "reason", "naturally selected", "path", m.Path)
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
			g.logger.Info("drop replace", "reason", "unused", "path", r.Old.Path)
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
			g.logger.Info("drop replace", "reason", "naturally selected", "path", m.Path)
			if err := g.EditDropreplace(m.Path); err != nil {
				return err
			}
		}
	}

	return nil
}

func (g *GomodHelper) ModDownload(path, version string) (*DownloadModule, error) {
	if g.downloadTemp == "" {
		temp, _ := ioutil.TempDir("", "gomod.*")
		g.downloadTemp = temp
	}
	run := g.goRunner.WithDir(g.downloadTemp)
	out, err := run.RunOutput("mod", "download", "-json", fmt.Sprintf("%s@%s", path, version))
	if err != nil {
		return nil, err
	}
	var mod DownloadModule
	if err := json.Unmarshal(out, &mod); err != nil {
		return nil, err
	}
	return &mod, nil
}

func (g *GomodHelper) EditRequire(path, version string) error {
	return g.edit(EditRequire, path, "", version)
}

func (g *GomodHelper) EditReplace(oldPath, newPath, version string) error {
	return g.edit(EditReplace, oldPath, newPath, version)
}

func (g *GomodHelper) EditDropreplace(path string) error {
	return g.edit(EditDropreplace, path, "", "")
}

func (g *GomodHelper) edit(verb ModeEditVerb, oldPath, newPath, version string) error {
	var arg string
	switch verb {
	case EditRequire:
		arg = fmt.Sprintf("-require=%s@%s", oldPath, version)
	case EditReplace:
		arg = fmt.Sprintf("-replace=%s=%s@%s", oldPath, newPath, version)
	case EditDropreplace:
		arg = fmt.Sprintf("-dropreplace=%s", oldPath)
	default:
		return fmt.Errorf("unsupported edit verb %v", verb)
	}

	g.logger.V(2).Info("go mod edit", "verb", verb, "oldPath", oldPath, "newPath", newPath, "version", version)
	_, err := g.goRunner.RunCombinedOutput("mod", "edit", "-fmt", arg)
	if err != nil {
		return err
	}
	return nil
}

func backupGomod(modfile string) string {
	dir, _ := ioutil.TempDir("", "gomod.*")
	file := path.Join(dir, "go.mod")

	data, _ := ioutil.ReadFile(modfile)
	ioutil.WriteFile(file, data, 0644) //nolint:errcheck
	return file
}

func restoreGomod(modfile, tempfile string) {
	data, _ := ioutil.ReadFile(tempfile)
	ioutil.WriteFile(modfile, data, 0644) //nolint:errcheck
}

func IsValidVersion(version string) bool {
	if version == "" || version == "v0.0.0" || version == "v0.0.0-00010101000000-000000000000" {
		return false
	}
	return true
}
