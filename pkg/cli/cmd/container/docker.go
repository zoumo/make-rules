package container

import (
	"fmt"
	"path"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/zoumo/golib/cli"

	"github.com/zoumo/make-rules/pkg/cli/cmd/utils"
	"github.com/zoumo/make-rules/pkg/cli/common"
	"github.com/zoumo/make-rules/pkg/git"
	"github.com/zoumo/make-rules/pkg/runner"
)

var (
	_ cli.Command        = &DockerBuildCommand{}
	_ cli.ComplexOptions = &DockerBuildCommand{}
)

type DockerBuildCommand struct {
	*common.CommonOptions

	dockerRunner *runner.Runner

	allTargets []string
	targets    []string
	git        *git.Repository
	version    string
}

func NewContainerBuildCommand() *cobra.Command {
	return cli.NewCobraCommand(&DockerBuildCommand{
		CommonOptions: common.NewCommonOptions(),
		dockerRunner:  runner.NewRunner("docker"),
	})
}

func (c *DockerBuildCommand) Name() string {
	return "build"
}

func (c *DockerBuildCommand) BindFlags(fs *pflag.FlagSet) {
	// Call embedded CommonOptions.BindFlags first
	c.CommonOptions.BindFlags(fs)

	fs.StringSliceVar(&c.Config.Container.Registries, "registries", c.Config.Container.Registries, "docker image registries")
	fs.StringVar(&c.version, "version", c.version, "go build target version")
}

func (c *DockerBuildCommand) Complete(cmd *cobra.Command, args []string) error {
	// Call embedded CommonOptions.Complete (sets Logger, loads Config)
	if err := c.CommonOptions.Complete(cmd, args); err != nil {
		return err
	}

	// no targets, walk cmd/ dir to find targets
	allTargets, err := utils.FindTargetsFrom(c.Workspace, "build", "Dockerfile")
	if err != nil {
		return err
	}
	c.allTargets = allTargets
	c.targets = utils.FilterTargets(args, c.allTargets, "build")

	r, err := git.Open(c.Workspace)
	if err != nil {
		c.Logger.Info("worksapce is not a git repo", "workspace", c.Workspace)
	} else {
		c.git = r
	}
	return nil
}

func (c *DockerBuildCommand) Validate() error {
	// Call embedded CommonOptions.Validate first
	if err := c.CommonOptions.Validate(); err != nil {
		return err
	}
	return nil
}

func (c *DockerBuildCommand) getDockerTag() string {
	version := "v0.0.0"
	if c.version != "" {
		return c.version
	}
	if c.git == nil {
		// this is not a git repo
		return version
	}

	// check dirty
	state, err := c.git.TreeState()
	if err != nil {
		c.Logger.Error(err, "failed to detect git tree state")
		return version
	}

	if c.version == "" {
		desc, err := c.git.Describe(nil)
		if err != nil {
			c.Logger.Error(err, "failed to describe git for HEAD")
			return version
		}
		version = desc.DokcerTag()
	}

	if state == git.GitTreeDirty {
		version += "-dirty"
	}
	return version
}

func (c *DockerBuildCommand) Run(cmd *cobra.Command, args []string) error {
	c.Logger.Info("=================================================")
	c.Logger.Info("Docker build", "targets", c.targets)
	for _, target := range c.targets {
		dockerfile := path.Join(c.Workspace, target, "Dockerfile")
		tag := fmt.Sprintf("%s%s%s:%s", c.Config.Container.ImagePrefix, path.Base(target), c.Config.Container.ImageSuffix, c.getDockerTag())
		c.Logger.Info("-------------------------------------------------")
		c.Logger.Info("Docker build", "dockerfile", dockerfile, "tag", tag)

		out, err := c.dockerRunner.RunCombinedOutput("build", "-f", dockerfile, "-t", tag, c.Workspace)
		if err != nil {
			c.Logger.Error(err, "failed to build image", "output", string(out))
			return err
		}
		if len(c.Config.Container.Registries) == 0 {
			continue
		}
		for _, r := range c.Config.Container.Registries {
			// new tag
			newTag := path.Join(r, tag)
			c.Logger.Info("Docker tag", "from", tag, "to", newTag)
			out, err = c.dockerRunner.RunCombinedOutput("tag", tag, newTag)
			if err != nil {
				c.Logger.Error(err, "failt to tag image", "output", string(out))
				return err
			}
		}
		// delete original tag
		c.Logger.Info("Docker remove image", "image", tag)
		out, err = c.dockerRunner.RunCombinedOutput("rmi", tag)
		if err != nil {
			c.Logger.Error(err, "failt to delete image", "output", string(out))
			return err
		}
	}
	c.Logger.Info("-------------------------------------------------")
	return nil
}
