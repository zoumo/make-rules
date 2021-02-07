package injection

import (
	"github.com/go-logr/logr"

	"github.com/zoumo/make-rules/pkg/config"
)

var _ RequiresConfig = &InjectionMixin{}
var _ RequiresLogger = &InjectionMixin{}
var _ RequiresWorkspace = &InjectionMixin{}

//nolint:golint
type InjectionMixin struct {
	Config    *config.Config
	Logger    logr.Logger
	Workspace string
}

func NewInjectionMixin() *InjectionMixin {
	return &InjectionMixin{
		Config: config.New(),
	}
}

func (m *InjectionMixin) InjectConfig(cfg *config.Config) {
	m.Config = cfg
}

func (m *InjectionMixin) InjectLogger(l logr.Logger) {
	m.Logger = l
}

func (m *InjectionMixin) InjectWorkspace(ws string) {
	m.Workspace = ws
}
