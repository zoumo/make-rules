package injection

import (
	"github.com/spf13/cobra"

	"github.com/zoumo/golib/cli/injection"
	"github.com/zoumo/golib/cli/plugin"

	"github.com/zoumo/make-rules/pkg/config"
)

var InjectLogger = injection.InjectLogger

var InjectWorkspace = injection.InjectWorkspace

func InjectConfig(cfg *config.Config) plugin.InitHook {
	return func(cmd *cobra.Command, sub plugin.Subcommand) error {
		if inject, ok := sub.(RequiresConfig); ok {
			inject.InjectConfig(cfg)
		}
		return nil
	}
}
