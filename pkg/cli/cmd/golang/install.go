package golang

import (
	"errors"
	"io/ioutil"
	"os"
	"path"
	"runtime"

	"github.com/spf13/cobra"
	"github.com/zoumo/golib/cli"

	"github.com/zoumo/make-rules/pkg/cli/common"
	"github.com/zoumo/make-rules/pkg/runner"
)

var _ cli.Command = &GoinstallCommand{}

type GoinstallCommand struct {
	*GobuildCommand
}

func NewGoInstallCommand() *cobra.Command {
	return cli.NewCobraCommand(&GoinstallCommand{
		GobuildCommand: &GobuildCommand{
			CommonOptions: common.NewCommonOptions(),
			goCmd:         runner.NewRunner("go"),
			bashCmd:       runner.NewRunner("bash"),
		},
	})
}

func (c *GoinstallCommand) Name() string {
	return "install"
}

func (c *GoinstallCommand) Run(cmd *cobra.Command, args []string) error {
	gobin := c.getGobinPath()
	if len(gobin) == 0 {
		return errors.New("failed to find GOBIN path, please set GOBIN or GOPATH in env")
	}

	local := platform{
		GOOS:   runtime.GOOS,
		GOARCH: runtime.GOARCH,
	}

	found := false
	for _, p := range c.platforms {
		if local == p {
			found = true
			break
		}
	}
	if !found {
		// skip
		c.Logger.Info("skip copying binaries, no output built for local platform")
		return nil
	}

	for _, target := range c.targets {
		outputFile := c.outputFile(local, target)
		data, err := ioutil.ReadFile(outputFile)
		if err != nil {
			return err
		}
		targetFile := path.Join(gobin, path.Base(outputFile))

		c.Logger.Info("copying", "from", outputFile, "to", targetFile)
		err = ioutil.WriteFile(targetFile, data, 0755)
		if err != nil {
			return err
		}
	}
	return nil
}

// find gobin from env GOBIN or GOPATH/bin
// if GOBIN and GOPATH is not set, return ""
func (c *GoinstallCommand) getGobinPath() string {
	gobin := os.Getenv("GOBIN")
	if len(gobin) != 0 {
		return gobin
	}

	gopath := os.Getenv("GOPATH")
	if len(gopath) != 0 {
		return path.Join(gopath, "bin")
	}
	return ""
}
