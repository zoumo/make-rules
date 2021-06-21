package injection

import (
	"github.com/zoumo/make-rules/pkg/config"
)

// RequiresConfig indicate the subcommand requires config
type RequiresConfig interface {
	InjectConfig(cfg *config.Config)
}
