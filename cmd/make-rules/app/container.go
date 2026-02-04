package app

import (
	"github.com/spf13/cobra"
	"github.com/zoumo/golib/log"

	"github.com/zoumo/make-rules/pkg/cli/cmd/container"
	"github.com/zoumo/make-rules/pkg/config"
)

var containerlogger = log.Log.WithName("container")

func newContainerCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:          "container",
		Short:        "Used to build container",
		SilenceUsage: true,
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := config.Load()
			if err != nil {
				cfg = &config.Config{}
			}
			containerlogger.V(1).Info("make-rules config", "config", cfg)
			return nil
		},
	}

	cmd.AddCommand(container.NewContainerBuildCommand())
	return cmd
}
