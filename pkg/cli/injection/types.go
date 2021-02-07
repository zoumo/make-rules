package injection

import (
	"github.com/go-logr/logr"

	"github.com/zoumo/make-rules/pkg/config"
)

// RequiresValidation indicate the subcommand requires a logger
type RequiresLogger interface {
	InjectLogger(obj logr.Logger)
}

// RequiresValidation indicate the subcommand requires workspace
type RequiresWorkspace interface {
	InjectWorkspace(ws string)
}

// RequiresConfig indicate the subcommand requires config
type RequiresConfig interface {
	InjectConfig(cfg *config.Config)
}
