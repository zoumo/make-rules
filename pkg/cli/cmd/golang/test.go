package golang

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/spf13/pflag"

	"github.com/zoumo/make-rules/pkg/cli/injection"
	"github.com/zoumo/make-rules/pkg/cli/plugin"
	"github.com/zoumo/make-rules/pkg/runner"
)

var (
	excludeTest = []string{
		"vendor",
		"hack",
		"scripts",
		"test",
		"tests",
		"testdata",
	}
)

type goUnittestSubcommand struct {
	*injection.InjectionMixin

	goCmd *runner.Runner

	allTests []string
}

func NewGoUnittestCommand() plugin.Subcommand {
	return &goUnittestSubcommand{
		InjectionMixin: injection.NewInjectionMixin(),
		goCmd:          runner.NewRunner("go"),
	}
}

func (c *goUnittestSubcommand) Name() string {
	return "unittest"
}

func (c *goUnittestSubcommand) BindFlags(fs *pflag.FlagSet) {

}

func (c *goUnittestSubcommand) PreRun(args []string) error {
	out, err := c.goCmd.RunOutput("list", "-test", "./...")
	if err != nil {
		c.Logger.Error(err, "failed to go list ./...", string(out))
		return err
	}

	allList := strings.Split(string(out), "\n")

	allTest := []string{}
	for _, t := range allList {
		if !strings.HasSuffix(t, ".test") {
			continue
		}
		allTest = append(allTest, t)
	}

	if len(allTest) == 0 {
		// no test target
		c.Logger.Info("not unit test target found")
		return nil
	}

	regs := []*regexp.Regexp{}
	c.Config.Go.Test.Exclude = append(c.Config.Go.Test.Exclude, excludeTest...)
	for _, e := range c.Config.Go.Test.Exclude {
		expr := fmt.Sprintf(".*/%s/?", e)
		reg, err := regexp.Compile(expr)
		if err != nil {
			c.Logger.Error(err, "invalid regexp", "expr", expr)
			continue
		}
		regs = append(regs, reg)
	}

	for _, test := range allTest {
		match := false
		for _, reg := range regs {
			if reg.MatchString(test) {
				match = true
				break
			}
		}
		if !match {
			c.allTests = append(c.allTests, test)
		}
	}

	return nil
}

func (c *goUnittestSubcommand) Run(args []string) error {
	for _, test := range c.allTests {
		test = strings.TrimSuffix(test, ".test")
		out, err := c.goCmd.RunCombinedOutput("test", test)
		if err != nil {
			c.Logger.Error(err, "unitest failed", "output", string(out))
			return err
		}
		c.Logger.Info("done", "package", test)
	}
	return nil
}
