package injection

import "github.com/go-logr/logr"

// RequiresValidation is a subcommand that requires injection
type RequiresInjection interface {
	InjectLogger(obj logr.Logger)
	InjectWorkspace(ws string)
}
