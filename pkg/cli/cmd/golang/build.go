package golang

import (
	"fmt"
	"os"
	"path"
	"strings"
	"time"

	"github.com/spf13/pflag"

	"github.com/zoumo/golib/cli/plugin"

	"github.com/zoumo/make-rules/pkg/cli/cmd/utils"
	"github.com/zoumo/make-rules/pkg/cli/injection"
	"github.com/zoumo/make-rules/pkg/git"
	"github.com/zoumo/make-rules/pkg/runner"
	"github.com/zoumo/make-rules/version"
)

var (
	RequiredGoEnvKeys = []string{
		"GO111MODULE",
		"GOFLAGS",
		"GOINSECURE",
		"GOMOD",
		"GOMODCACHE",
		"GONOPROXY",
		"GONOSUMDB",
		"GOPATH",
		"GOPROXY",
		"GOROOT",
		"GOSUMDB",
	}
)

type platform struct {
	GOOS   string
	GOARCH string
}

func (p platform) String() string {
	return p.GOOS + "/" + p.GOARCH
}

func readPlatform(val string) *platform {
	ps := strings.SplitN(val, "/", 2)
	if len(ps) != 2 {
		return nil
	}
	return &platform{GOOS: ps[0], GOARCH: ps[1]}
}

type gobuildSubcommand struct {
	*injection.InjectionMixin

	goCmd   *runner.Runner
	bashCmd *runner.Runner

	allTargets []string
	targets    []string
	platforms  []platform
	module     string

	git     *git.Repository
	version string
}

func NewGobuildCommand() plugin.Subcommand {
	return &gobuildSubcommand{
		InjectionMixin: injection.NewInjectionMixin(),
		goCmd:          runner.NewRunner("go"),
		bashCmd:        runner.NewRunner("bash"),
	}
}

func (c *gobuildSubcommand) Name() string {
	return "build"
}

func (c *gobuildSubcommand) BindFlags(fs *pflag.FlagSet) {
	// convert platforms
	c.Config.SetDefaults()

	fs.StringSliceVar(&c.Config.Go.Build.Platforms, "platforms", c.Config.Go.Build.Platforms, "go build target platforms")
	fs.StringVar(&c.Config.Go.Build.GlobalHooksDir, "global-hooks-dir", c.Config.Go.Build.GlobalHooksDir, "the global pre-build and post-build hooks dir")
	fs.StringVar(&c.version, "version", c.version, "go build target version")
}

func (c *gobuildSubcommand) init(args []string) error {
	for _, p := range c.Config.Go.Build.Platforms {
		if pf := readPlatform(p); pf != nil {
			c.platforms = append(c.platforms, *pf)
		}
	}

	// find module
	out, err := c.goCmd.RunOutput("list", "-m")
	if err != nil {
		c.Logger.Error(err, string(out))
		return err
	}
	c.module = strings.TrimSpace(string(out))
	// no targets, walk cmd/ dir to find targets
	allTargets, err := utils.FindTargetsFrom(c.Workspace, "cmd", "main.go")
	if err != nil {
		return err
	}
	c.allTargets = allTargets
	c.targets = utils.FilterTargets(args, c.allTargets, "cmd")

	if len(c.targets) == 0 {
		c.Logger.Info("!! no valid go build target specified")
		return nil
	}

	r, err := git.Open(c.Workspace)
	if err != nil {
		c.Logger.Info("worksapce is not a git repo", "workspace", c.Workspace)
	} else {
		c.git = r
	}

	c.Logger.Info("Before Go compiling", "platforms", c.Config.Go.Build.Platforms, "targets", c.targets)
	return nil
}

func (c *gobuildSubcommand) getVersionInfo() version.Info {
	info := version.Info{
		GitVersion:   "v0.0.0",
		GitCommit:    "unknown",
		GitTreeState: "unknown",
		GitRemote:    "unknown",
		BuildDate:    time.Now().UTC().Format(time.RFC3339),
	}

	if c.version != "" {
		// use user defined version
		info.GitVersion = c.version
	}

	if c.git == nil {
		// this is not a git repo
		return info
	}

	head, err := c.git.Head()
	if err != nil {
		c.Logger.Error(err, "failed to get head of git repo")
		return info
	}
	info.GitCommit = head.Hash().String()

	state, err := c.git.TreeState()
	if err != nil {
		c.Logger.Error(err, "failed to detect git tree state")
		return info
	}
	info.GitTreeState = string(state)

	if c.version == "" {
		// get version from git describe
		desc, err := c.git.Describe(head)
		if err != nil {
			c.Logger.Error(err, "failed to describe git for HEAD")
			return info
		}
		info.GitVersion = desc.SemanticVersion()
	}

	if info.GitTreeState == string(git.GitTreeDirty) {
		info.GitVersion += "-dirty"
	}

	remote, err := c.git.RemoteURL("origin")
	if err != nil {
		c.Logger.Error(err, "failed to get origin remote url from git")
		return info
	}
	info.GitRemote = remote
	return info
}

func (c *gobuildSubcommand) ldflags() string {
	info := c.getVersionInfo()

	flags := []string{
		fmt.Sprintf("-X github.com/zoumo/make-rules/version.buildDate=%s", info.BuildDate),
		fmt.Sprintf("-X github.com/zoumo/make-rules/version.gitVersion=%s", info.GitVersion),
		fmt.Sprintf("-X github.com/zoumo/make-rules/version.gitCommit=%s", info.GitCommit),
		fmt.Sprintf("-X github.com/zoumo/make-rules/version.gitRemote=%s", info.GitRemote),
		fmt.Sprintf("-X github.com/zoumo/make-rules/version.gitTreeState=%s", info.GitTreeState),
	}

	flags = append(flags, c.Config.Go.Build.LDFlags...)

	return strings.Join(flags, " ")
}

type HookPhase string

const (
	PreBuildHook  = "pre-build"
	PostBuildHook = "post-build"
)

func (c *gobuildSubcommand) runHook(dir string, phase HookPhase) error {
	// run global hooks
	hook := path.Join(dir, string(phase))
	c.Logger.V(2).Info("detecting hook", "path", hook)
	if _, err := os.Stat(hook); err == nil {
		outdirs := []string{}
		for _, p := range c.platforms {
			outdirs = append(outdirs, path.Dir(c.outputFile(p, "nothing")))
		}
		cmd := c.bashCmd.WithEnvs(
			"MAKE_RULES_WORKSPACE", c.Workspace,
			"MAKE_RULES_GO_BUILD_BINARY_DIRS", strings.Join(outdirs, ","),
			"MAKE_RULES_GO_BUILD_PLATFORMS", strings.Join(c.Config.Go.Build.Platforms, ","),
		)
		c.Logger.Info("hook started", "phase", phase, "path", hook)
		out, err := cmd.RunCombinedOutput(hook)
		if err != nil {
			c.Logger.Error(err, string(out))
			return err
		}
		c.Logger.Info("hook completed", "phase", phase, "output", string(out))
	}
	return nil
}

func (c *gobuildSubcommand) PreRun(args []string) error {
	if err := c.init(args); err != nil {
		return err
	}
	// run global hooks
	if err := c.runHook(c.Config.Go.Build.GlobalHooksDir, PreBuildHook); err != nil {
		return err
	}
	return nil
}

func (c *gobuildSubcommand) PostRun(args []string) error {
	// run global hooks
	if err := c.runHook(c.Config.Go.Build.GlobalHooksDir, PostBuildHook); err != nil {
		return err
	}
	return nil
}

func (c *gobuildSubcommand) Run(args []string) error {
	for _, platform := range c.platforms {
		c.Logger.Info("Go cross compiling", "platform", platform.String())
		cmd := c.goCmd.WithEnvs(
			"GOOS", platform.GOOS,
			"GOARCH", platform.GOARCH,
			"WORKSPACE", c.Workspace,
		)
		for _, t := range c.targets {
			target := path.Join(c.module, t)
			output := c.outputFile(platform, target)
			hookDir := path.Join(c.Workspace, t)
			// run pre build hook
			if err := c.runHook(hookDir, PreBuildHook); err != nil {
				return err
			}
			c.Logger.Info("Go build started", "module", target, "output", output)
			// go build
			args := c.gobuildArgs(output, target)
			if c.Logger.V(2).Enabled() {
				kvlist := []interface{}{}
				for k, v := range cmd.FilterEnv(RequiredGoEnvKeys) {
					kvlist = append(kvlist, k, v)
				}
				kvlist = append(kvlist, "args", args)
				c.Logger.Info("Go env and args", kvlist...)
			}
			out, err := cmd.RunCombinedOutput(args...)
			if err != nil {
				c.Logger.Error(err, string(out))
				return err
			}
			c.Logger.Info("Go build completed", "module", target)
			// run post build hook
			if err := c.runHook(hookDir, PostBuildHook); err != nil {
				return err
			}
		}
	}
	return nil
}

func (c gobuildSubcommand) gobuildArgs(output, target string) []string {
	args := []string{"build", "-i"}
	if len(c.Config.Go.Build.Flags) > 0 {
		args = append(args, c.Config.Go.Build.Flags...)
	}
	args = append(args, "-ldflags", c.ldflags())
	if len(c.Config.Go.Build.GCFlags) > 0 {
		args = append(args, "-gcflags", strings.Join(c.Config.Go.Build.GCFlags, " "))
	}
	args = append(args, "-o", output, target)
	return args
}

func (c *gobuildSubcommand) outputFile(platform platform, target string) string {
	bin := path.Base(target)
	if platform.GOOS == "windown" {
		bin += ".exe"
	}
	return fmt.Sprintf("%s/bin/%s_%s/%s", c.Workspace, platform.GOOS, platform.GOARCH, bin)
}
