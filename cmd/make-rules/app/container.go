package app

import (
	"github.com/spf13/cobra"

	"github.com/zoumo/golib/cli/plugin"

	"github.com/zoumo/make-rules/pkg/cli/cmd/container"
	"github.com/zoumo/make-rules/pkg/cli/injection"
	"github.com/zoumo/make-rules/pkg/config"
	"github.com/zoumo/make-rules/pkg/log"
)

var (
	containerlogger = log.Log.WithName("container")
)

func newContainerCommand(cfg *config.Config) *cobra.Command {
	cmd := &cobra.Command{
		Use:          "container",
		Short:        "Used to build container",
		SilenceUsage: true,
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			containerlogger.V(1).Info("make-rules config", "config", cfg)
			return nil
		},
	}

	cmd.AddCommand(plugin.NewCobraSubcommandOrDie(
		container.NewContainerBuildCommand(),
		injection.InjectLogger(containerlogger.WithName("build")),
		injection.InjectWorkspace(),
		injection.InjectConfig(cfg),
	))
	return cmd
}
