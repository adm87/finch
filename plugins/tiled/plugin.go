package tiled

import (
	"github.com/adm87/finch/pkg/application"
	"github.com/adm87/finch/pkg/errs"
)

type Plugin struct {
	config *Config
}

func NewPlugin(cfg *Config) *Plugin {
	return &Plugin{config: cfg}
}

func (p *Plugin) Name() string {
	return "tiled"
}

// Init checks the config. It will also register the .tmx and .tsx importers and
// the map renderer.
func (p *Plugin) Init(*application.Application) error {
	if p.config == nil {
		return errs.ArgumentNil("config")
	}
	return p.config.Validate()
}
