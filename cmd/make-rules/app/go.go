package app

import (
	"github.com/spf13/cobra"
	"github.com/zoumo/golib/log"

	"github.com/zoumo/make-rules/pkg/cli/cmd/golang"
	"github.com/zoumo/make-rules/pkg/config"
	goutil "github.com/zoumo/make-rules/pkg/golang"
)

var (
	gologger = log.Log.WithName("go")
)

func newGoCommand(cfg *config.Config) *cobra.Command {
	cmd := &cobra.Command{
		Use:          "go",
		Short:        "Used to build go module and operate go.mod",
		SilenceUsage: true,
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			gologger.V(1).Info("make-rules config", "config", cfg)
			if err := goutil.VerifyGoVersion(cfg.Go.MinimumVersion); err != nil {
				return err
			}
			return nil
		},
	}
	cmd.AddCommand(golang.NewGobuildCommand())
	cmd.AddCommand(golang.NewGoInstallCommand())
	cmd.AddCommand(newGoModCommand())
	cmd.AddCommand(golang.NewFormatSubcommand())
	cmd.AddCommand(golang.NewGoUnittestCommand())

	return cmd
}

func newGoModCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:          "mod",
		Short:        "Manage go.mod",
		SilenceUsage: true,
	}
	cmd.AddCommand(golang.NewModTidyCommand())
	cmd.AddCommand(golang.NewModRequireCommand())
	cmd.AddCommand(golang.NewModReplaceCommand())
	cmd.AddCommand(golang.NewModUpdateCommand())
	return cmd
}
