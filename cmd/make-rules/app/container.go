package app

import (
	"github.com/spf13/cobra"

	"github.com/zoumo/make-rules/pkg/cli/cmd/container"
	"github.com/zoumo/make-rules/pkg/cli/injection"
	"github.com/zoumo/make-rules/pkg/cli/plugin"
	"github.com/zoumo/make-rules/pkg/log"
)

var (
	containerlogger = log.Log.WithName("container")
)

func newContainerCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:          "container",
		Short:        "Used to build container",
		SilenceUsage: true,
	}
	cmd.AddCommand(plugin.NewCobraSubcommandOrDie(
		container.NewContainerBuildCommand(),
		injection.InjectLogger(containerlogger.WithName("build")),
		injection.InjectWorkspace(),
	))
	return cmd
}
