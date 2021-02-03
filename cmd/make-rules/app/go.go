package app

import (
	"github.com/spf13/cobra"

	"github.com/zoumo/make-rules/pkg/cli/cmd/golang"
	"github.com/zoumo/make-rules/pkg/cli/injection"
	"github.com/zoumo/make-rules/pkg/cli/plugin"
	goutil "github.com/zoumo/make-rules/pkg/golang"
	"github.com/zoumo/make-rules/pkg/log"
)

var (
	gologger = log.Log.WithName("go")
)

func newGoCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:          "go",
		Short:        "Used to build go module and operate go.mod",
		SilenceUsage: true,
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			if err := goutil.VerifyGoVersion(); err != nil {
				return err
			}
			return nil
		},
	}
	cmd.AddCommand(newGoBuildCommand())
	cmd.AddCommand(newGoModCommand())

	return cmd
}

func newGoBuildCommand() *cobra.Command {
	return plugin.NewCobraSubcommandOrDie(
		golang.NewGobuildCommand(),
		injection.InjectLogger(gologger.WithName("build")),
		injection.InjectWorkspace(),
	)
}

func newGoModCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:          "mod",
		SilenceUsage: true,
	}
	cmd.AddCommand(plugin.NewCobraSubcommandOrDie(
		golang.NewModTidyCommand(),
		injection.InjectLogger(gologger.WithName("mod")),
		injection.InjectWorkspace(),
	))
	cmd.AddCommand(plugin.NewCobraSubcommandOrDie(
		golang.NewModRequireCommand(),
		injection.InjectLogger(gologger.WithName("mod")),
		injection.InjectWorkspace(),
	))
	return cmd
}
