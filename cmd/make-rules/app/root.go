package app

import (
	"github.com/spf13/cobra"
	"github.com/zoumo/golib/log/consolog"

	cliflag "github.com/zoumo/make-rules/pkg/cli/flag"
	"github.com/zoumo/make-rules/pkg/config"
	"github.com/zoumo/make-rules/pkg/log"
	"github.com/zoumo/make-rules/version"
)

func init() {
	log.SetLogger(consolog.New())
}

func NewRootCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:          "make-rules",
		SilenceUsage: true,
	}

	// add global flags
	cmd.SetGlobalNormalizationFunc(cliflag.WordSepNormalizeFunc)
	cliflag.AddGlobalFlags(cmd.PersistentFlags())

	cfg, err := config.Load()
	if err != nil {
		cfg = &config.Config{}
	}

	// add subcommand
	cmd.AddCommand(newGoCommand(cfg))
	cmd.AddCommand(newContainerCommand(cfg))
	cmd.AddCommand(version.NewCommand())

	return cmd
}
