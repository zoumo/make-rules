package injection

import (
	"github.com/go-logr/logr"
)

//nolint:golint
type InjectionMixin struct {
	Logger    logr.Logger
	Workspace string
}

func (m *InjectionMixin) InjectLogger(l logr.Logger) {
	m.Logger = l
}

func (m *InjectionMixin) InjectWorkspace(ws string) {
	m.Workspace = ws
}
