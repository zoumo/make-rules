package container

import (
	"fmt"
	"path"

	"github.com/spf13/pflag"

	"github.com/zoumo/make-rules/pkg/cli/cmd/utils"
	"github.com/zoumo/make-rules/pkg/cli/injection"
	"github.com/zoumo/make-rules/pkg/cli/plugin"
	"github.com/zoumo/make-rules/pkg/git"
	"github.com/zoumo/make-rules/pkg/runner"
)

type dockerBuildSubcommand struct {
	*injection.InjectionMixin

	dockerRunner *runner.Runner

	registries []string
	allTargets []string
	targets    []string
	git        *git.Repository
	version    string
}

func NewContainerBuildCommand() plugin.Subcommand {
	return &dockerBuildSubcommand{
		InjectionMixin: &injection.InjectionMixin{},
		dockerRunner:   runner.NewRunner("docker"),
	}
}

func (c *dockerBuildSubcommand) Name() string {
	return "build"
}

func (c *dockerBuildSubcommand) BindFlags(fs *pflag.FlagSet) {
	fs.StringSliceVar(&c.registries, "registries", c.registries, "docker image registries")
	fs.StringVar(&c.version, "version", c.version, "go build target version")
}

func (c *dockerBuildSubcommand) PreRun(args []string) error {
	// no targets, walk cmd/ dir to find targets
	allTargets, err := utils.FindTargetsFrom(c.Workspace, "build", "Dockerfile")
	if err != nil {
		return err
	}
	c.allTargets = allTargets
	c.targets = utils.FilterTargets(args, c.allTargets, "build")
	for i := range c.targets {
		// nomalize target path
		c.targets[i] = path.Join(c.Workspace, c.targets[i])
	}

	r, err := git.Open(c.Workspace)
	if err != nil {
		c.Logger.Info("worksapce is not a git repo", "workspace", c.Workspace)
	} else {
		c.git = r
	}
	return nil
}

func (c *dockerBuildSubcommand) getDockerTag() string {
	version := "v0.0.0"
	if c.version != "" {
		version = c.version
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

func (c *dockerBuildSubcommand) Run(args []string) error {
	c.Logger.Info("=================================================")
	c.Logger.Info("Docker build", "targets", c.targets)
	for _, target := range c.targets {
		dockerfile := path.Join(target, "Dockerfile")
		tag := fmt.Sprintf("%s:%s", path.Base(target), c.getDockerTag())
		c.Logger.Info("-------------------------------------------------")
		c.Logger.Info("Docker build", "dockerfile", dockerfile, "tag", tag)

		out, err := c.dockerRunner.RunCombinedOutput("build", "-f", dockerfile, "-t", tag, c.Workspace)
		if err != nil {
			c.Logger.Error(err, "failed to build image", "output", string(out))
			return err
		}
		if len(c.registries) == 0 {
			continue
		}
		for _, r := range c.registries {
			// new tag
			newTag := fmt.Sprintf("%s/%s", r, tag)
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
