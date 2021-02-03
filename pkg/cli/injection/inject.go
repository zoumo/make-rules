package injection

import (
	"os"

	"github.com/go-logr/logr"
	"github.com/spf13/cobra"

	"github.com/zoumo/make-rules/pkg/cli/plugin"
)

func InjectLogger(logger logr.Logger) plugin.InitHook {
	return func(cmd *cobra.Command, sub plugin.Subcommand) error {
		if injection, ok := sub.(RequiresInjection); ok {
			injection.InjectLogger(logger)
		}
		return nil
	}
}

func InjectWorkspace() plugin.InitHook {
	return func(cmd *cobra.Command, sub plugin.Subcommand) error {
		ws, err := os.Getwd()
		if err != nil {
			return err
		}
		if injection, ok := sub.(RequiresInjection); ok {
			injection.InjectWorkspace(ws)
		}
		return nil
	}
}
