package app

import (
	"github.com/spf13/cobra"
	"github.com/zoumo/golib/cli/plugin"

	"github.com/zoumo/make-rules/pkg/cli/cmd/golang"
	"github.com/zoumo/make-rules/pkg/cli/injection"
	"github.com/zoumo/make-rules/pkg/config"
	goutil "github.com/zoumo/make-rules/pkg/golang"
	"github.com/zoumo/make-rules/pkg/log"
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
			if err := goutil.VerifyGoVersion(); err != nil {
				return err
			}
			return nil
		},
	}
	cmd.AddCommand(newGoBuildCommand(cfg))
	cmd.AddCommand(newGoInstallCommand(cfg))
	cmd.AddCommand(newGoModCommand(cfg))
	cmd.AddCommand(newGoFormatCommand(cfg))
	cmd.AddCommand(newGoUnittestCommand(cfg))

	return cmd
}

func newGoBuildCommand(cfg *config.Config) *cobra.Command {
	return plugin.NewCobraSubcommandOrDie(
		golang.NewGobuildCommand(),
		injection.InjectLogger(gologger.WithName("build")),
		injection.InjectWorkspace(),
		injection.InjectConfig(cfg),
	)
}

func newGoInstallCommand(cfg *config.Config) *cobra.Command {
	return plugin.NewCobraSubcommandOrDie(
		golang.NewGoInstallCommand(),
		injection.InjectLogger(gologger.WithName("build")),
		injection.InjectWorkspace(),
		injection.InjectConfig(cfg),
	)
}

func newGoFormatCommand(cfg *config.Config) *cobra.Command {
	return plugin.NewCobraSubcommandOrDie(
		golang.NewFormatSubcommand(),
		injection.InjectLogger(gologger.WithName("format")),
		injection.InjectWorkspace(),
		injection.InjectConfig(cfg),
	)
}

func newGoModCommand(cfg *config.Config) *cobra.Command {
	cmd := &cobra.Command{
		Use:          "mod",
		SilenceUsage: true,
	}
	logger := gologger.WithName("mod")
	cmd.AddCommand(plugin.NewCobraSubcommandOrDie(
		golang.NewModTidyCommand(),
		injection.InjectLogger(logger),
		injection.InjectWorkspace(),
		injection.InjectConfig(cfg),
	))
	cmd.AddCommand(plugin.NewCobraSubcommandOrDie(
		golang.NewModRequireCommand(),
		injection.InjectLogger(logger),
		injection.InjectWorkspace(),
		injection.InjectConfig(cfg),
	))
	cmd.AddCommand(plugin.NewCobraSubcommandOrDie(
		golang.NewModReplaceCommand(),
		injection.InjectLogger(logger),
		injection.InjectWorkspace(),
		injection.InjectConfig(cfg),
	))
	cmd.AddCommand(plugin.NewCobraSubcommandOrDie(
		golang.NewModUpdateCommand(),
		injection.InjectLogger(logger),
		injection.InjectWorkspace(),
		injection.InjectConfig(cfg),
	))
	return cmd
}

func newGoUnittestCommand(cfg *config.Config) *cobra.Command {
	return plugin.NewCobraSubcommandOrDie(
		golang.NewGoUnittestCommand(),
		injection.InjectLogger(gologger.WithName("unittest")),
		injection.InjectWorkspace(),
		injection.InjectConfig(cfg),
	)
}
