package tiled

import (
	"github.com/adm87/finch/pkg/application"
)

type Plugin struct {
	config *Config
	maps   []MapConfig
}

func NewPlugin(cfg *Config) *Plugin {
	return &Plugin{config: cfg}
}

func (p *Plugin) Name() string {
	return "tiled"
}

// Init will register the .tmx and .tsx importers and the map renderer.
func (p *Plugin) Init(*application.Application) error {
	return nil
}
