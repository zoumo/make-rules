package golang

import (
	"bufio"
	"bytes"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/go-logr/logr"
	"github.com/spf13/pflag"

	"github.com/zoumo/make-rules/pkg/cli/injection"
	"github.com/zoumo/make-rules/pkg/cli/plugin"
	"github.com/zoumo/make-rules/pkg/runner"
)

var (
	excludeDir = []string{
		".git",
		"vendor",
		"hack",
		"bin",
		"temp",
		"output",
		"generated",
	}
	excludeFileRegexps = []string{
		".*generated.*",
	}
)

type exclude struct {
	files       map[string]bool
	dirs        map[string]bool
	fileRegexps []*regexp.Regexp
	dirRegexps  []*regexp.Regexp
	logger      logr.Logger
}

func defaultExclued(logger logr.Logger) *exclude {
	e := &exclude{
		files:       make(map[string]bool),
		dirs:        make(map[string]bool),
		fileRegexps: make([]*regexp.Regexp, 0),
		dirRegexps:  make([]*regexp.Regexp, 0),
		logger:      logger,
	}

	for _, f := range excludeDir {
		e.AddDir(f)
	}
	for _, f := range excludeFileRegexps {
		e.AddFileRegexp(f)
	}
	return e
}

func (e *exclude) MatchDir(dir string) (string, bool) {
	basename := path.Base(dir)
	if _, ok := e.dirs[basename]; ok {
		return basename, true
	}
	for _, reg := range e.dirRegexps {
		if reg.MatchString(dir) {
			return reg.String(), true
		}
	}
	return "", false
}

func (e *exclude) MatchFile(filename string) (string, bool) {
	basename := path.Base(filename)
	if _, ok := e.files[basename]; ok {
		return basename, true
	}
	for _, reg := range e.fileRegexps {
		if reg.MatchString(filename) {
			return reg.String(), true
		}
	}
	return "", false
}

// Add a explicit dir path not a regexp
func (e *exclude) AddDir(dir string) {
	e.dirs[path.Base(dir)] = true
}

// Add a dir path regexp
func (e *exclude) AddDirRegexp(dirE string) {
	// firstly we add dir regexp directly to try to match the base path
	e.dirs[dirE] = true
	// compile regexp
	reg, err := regexp.Compile(dirE)
	if err != nil {
		e.logger.Error(err, "invalid regexp expression", "expr", dirE)
		return
	}
	e.dirRegexps = append(e.dirRegexps, reg)
}

func (e *exclude) AddFile(file string) {
	e.files[path.Base(file)] = true
}

func (e *exclude) AddFileRegexp(fileE string) {
	e.files[fileE] = true
	// compile regexp
	reg, err := regexp.Compile(fileE)
	if err != nil {
		e.logger.Error(err, "invalid regexp expression", "expr", fileE)
		return
	}
	e.fileRegexps = append(e.fileRegexps, reg)
}

type formatSubcommand struct {
	*injection.InjectionMixin
	goCmd        *runner.Runner
	goimportsCmd *runner.Runner

	module string
}

func NewFormatSubcommand() plugin.Subcommand {
	return &formatSubcommand{
		InjectionMixin: injection.NewInjectionMixin(),
		goCmd:          runner.NewRunner("go"),
		goimportsCmd:   runner.NewRunner("goimports"),
	}
}

func (c *formatSubcommand) Name() string {
	return "format"
}

func (c *formatSubcommand) BindFlags(fs *pflag.FlagSet) {
}

func (c *formatSubcommand) PreRun(args []string) error {
	c.Config.SetDefaults()
	// find module
	out, err := c.goCmd.RunOutput("list", "-m")
	if err != nil {
		c.Logger.Error(err, string(out))
		return err
	}
	c.module = strings.TrimSpace(string(out))
	if c.Config.Go.Format.Local == "" {
		c.Config.Go.Format.Local = c.module
	}
	return nil
}

func (c *formatSubcommand) Run(args []string) error {
	if len(args) > 0 {
		// format used defined targets
		for _, f := range args {
			if err := c.format(f); err != nil {
				return err
			}
		}
		return nil
	}

	exclude := defaultExclued(c.Logger)
	for _, e := range c.Config.Go.Format.Exclude.Dirs {
		exclude.AddDirRegexp(e)
	}
	for _, e := range c.Config.Go.Format.Exclude.Files {
		exclude.AddFileRegexp(e)
	}

	// regexp.Compile()
	return filepath.Walk(c.Workspace, func(file string, info os.FileInfo, err error) error {
		if info.IsDir() {
			// skip some dir
			rule, match := exclude.MatchDir(file)
			if match {
				c.Logger.Info("skip", "dir", file, "matchRule", rule)
				return filepath.SkipDir
			}
			return nil
		}
		// file

		// skip some file
		if !strings.HasSuffix(file, ".go") {
			// skip not go file
			return nil
		}
		rule, match := exclude.MatchFile(file)
		if match {
			c.Logger.Info("skip", "file", file, "matchRule", rule)
			return nil
		}

		// format
		return c.format(file)
	})
}

func (c *formatSubcommand) format(finename string) error {
	// format
	// delete empty line between import ( and )
	if err := deleteEmptyLineWithinImports(finename); err != nil {
		c.Logger.Error(err, "failed to delete empty line with in import()")
		return err
	}
	// format imports
	if out, err := c.goimportsCmd.RunCombinedOutput("-format-only", "-w", "-local", c.Config.Go.Format.Local, finename); err != nil {
		c.Logger.Error(err, "failed to format go file", "file", finename, "output", string(out))
		return err
	}
	c.Logger.Info("done", "file", finename)
	return nil
}

func deleteEmptyLineWithinImports(filename string) error {
	f, err := os.Open(filename)
	if err != nil {
		return err
	}

	s := bufio.NewScanner(f)
	newfile := &bytes.Buffer{}
	inImport := false
	for s.Scan() {
		text := s.Text()
		trimed := strings.TrimSpace(text)
		if trimed == "import (" {
			inImport = true
		} else if inImport {
			if trimed == ")" {
				inImport = false
			} else if trimed == "" {
				continue
			}
		}

		newfile.WriteString(text + "\n")
	}
	if s.Err() != nil {
		return s.Err()
	}
	return ioutil.WriteFile(filename, newfile.Bytes(), 0644)
}
