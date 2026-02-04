package common

import (
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/zoumo/golib/cli"

	"github.com/zoumo/make-rules/pkg/config"
)

// CommonOptions provides common options for all commands.
// It combines golib's cli.CommonOptions with project-specific config.
// Following golib/cli pattern: commands embed this struct and implement Command interface.
type CommonOptions struct {
	*cli.CommonOptions // Provides Logger and Workspace fields
	Config             *config.Config
}

// BindFlags implements cli.Options interface.
// MUST call embedded CommonOptions.BindFlags first to bind workspace/v flags.
func (o *CommonOptions) BindFlags(fs *pflag.FlagSet) {
	o.CommonOptions.BindFlags(fs)
}

// Complete implements cli.ComplexOptions interface.
// MUST call embedded CommonOptions.Complete to initialize Logger.
func (o *CommonOptions) Complete(cmd *cobra.Command, args []string) error {
	if err := o.CommonOptions.Complete(cmd, args); err != nil {
		return err
	}

	// Load config if not set
	if o.Config == nil {
		cfg, err := config.Load()
		if err != nil {
			o.Config = config.New()
		} else {
			o.Config = cfg
		}
	}
	o.Config.SetDefaults()

	return nil
}

// Validate implements cli.ComplexOptions interface.
// MUST call embedded CommonOptions.Validate first.
func (o *CommonOptions) Validate() error {
	return o.CommonOptions.Validate()
}

// NewCommonOptions creates a new CommonOptions with initialized config.
func NewCommonOptions() *CommonOptions {
	return &CommonOptions{
		CommonOptions: &cli.CommonOptions{},
		Config:        config.New(),
	}
}
