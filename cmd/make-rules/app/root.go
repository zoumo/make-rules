package app

import (
	"github.com/spf13/cobra"
	"github.com/zoumo/golib/log"
	"github.com/zoumo/golib/log/consolog"

	cliflag "github.com/zoumo/make-rules/pkg/cli/flag"
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

	// add subcommand
	cmd.AddCommand(newGoCommand())
	cmd.AddCommand(newContainerCommand())
	cmd.AddCommand(version.NewCommand())

	return cmd
}
