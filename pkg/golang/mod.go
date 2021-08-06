package golang

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"path"
	"strings"

	"github.com/Masterminds/semver/v3"
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
	pinned       goset.Set
	goVersion    *semver.Version
}

func NewGomodHelper(modfile string, logger logr.Logger) *GomodHelper {
	version, err := getGoVersion()
	if err != nil {
		panic(fmt.Sprintf("failed to get go version: %v", err))
	}
	g := &GomodHelper{
		modfile:   modfile,
		goRunner:  runner.NewRunner("go").WithDir(path.Dir(modfile)),
		logger:    logger,
		pinned:    goset.NewSet(),
		goVersion: version,
	}

	return g
}

func (g *GomodHelper) Require(path, version string, skipDeps bool) error {
	g.logger.Info("mod require", "path", path, "version", version, "skip-deps", skipDeps)
	mod, err := g.ModDownload(path, version)
	if err != nil {
		return err
	}

	// copy modfile to temp
	modfile := backupGomod(mod.GoMod)
	if modfile == "" {
		return fmt.Errorf("module %s version %s does not have a go.mod file", path, version)
	}

	g.logger.Info("download required package/go.mod to temp dir", "package", path, "source", mod.GoMod, "target", modfile)

	// replace version with mod version
	version = mod.Version

	newgomod := NewGomodHelper(modfile, g.logger)

	// kubernetes imports module in its staging path
	// we must replace these module firstly to avoid error occurring
	if path == "k8s.io/kubernetes" {
		gomod, err := newgomod.ParseMod()
		if err != nil {
			return err
		}
		if err := g.replaceK8SStaging(gomod, version); err != nil {
			return err
		}
	}

	if skipDeps {
		return g.pinDependence(path, path, version)
	}

	// find local list deps
	requireMod, err := newgomod.ParseMod()
	if err != nil {
		return err
	}

	for _, r := range requireMod.Require {
		if IsValidVersion(r.Version) {
			g.logger.V(2).Info(fmt.Sprintf("require %s@%s", r.Path, r.Version), "source", "require")
			if err := g.EditRequire(r.Path, r.Version); err != nil {
				return err
			}
		}
	}

	for _, r := range requireMod.Replace {
		if IsValidVersion(r.New.Version) {
			g.logger.V(2).Info(fmt.Sprintf("require %s@%s", r.Old.Path, r.New.Version), "source", "replace")
			if err := g.EditRequire(r.Old.Path, r.New.Version); err != nil {
				return err
			}
		}
	}

	return g.pinDependence(path, path, version)
}

func (g *GomodHelper) Replace(oldPath, newPath, version string) error {
	if !IsValidVersion(version) {
		// replace to local path using a invalid version
		return g.pinDependence(oldPath, newPath, "v0.0.0")
	}

	// find true version
	mod, err := g.ModDownload(newPath, version)
	if err != nil {
		return err
	}
	version = mod.Version
	return g.pinDependence(oldPath, newPath, version)
}

func (g *GomodHelper) pinDependence(oldPath, newPath, version string) error {
	g.logger.Info("pin dependence", "oldPath", oldPath, "newPath", newPath, "version", version)
	if err := g.EditRequire(oldPath, version); err != nil {
		return err
	}
	if err := g.EditReplace(oldPath, newPath, version); err != nil {
		return err
	}

	g.pinned.Add(oldPath) //nolint
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

	if g.goVersion.Compare(go1160) >= 0 {
		// go1.16.0
		// we must run go mod tidy before go list if go version is
		// greater than go1.16.0, otherwise it will fail.
		err := g.ModTidy()
		if err != nil {
			return nil, err
		}
	}
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
		if IsValidVersion(version) {
			arg = fmt.Sprintf("-replace=%s=%s@%s", oldPath, newPath, version)
		} else {
			arg = fmt.Sprintf("-replace=%s=%s", oldPath, newPath)
		}
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
	switch version {
	case "", "v0.0.0", "v0.0.0-00010101000000-000000000000", "v0":
		return false
	}
	return true
}
