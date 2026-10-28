package aseprite

import (
	"github.com/adm87/finch/pkg/application"
)

// Plugin builds the Aseprite libraries from the config the game gives it.
type Plugin struct {
	libraries *Libraries
	config    *Config
}

func NewPlugin(cfg *Config) *Plugin {
	return &Plugin{config: cfg}
}

func (p *Plugin) Name() string {
	return "aseprite"
}

// Libraries is nil until Init runs.
func (p *Plugin) Libraries() *Libraries {
	return p.libraries
}

func (p *Plugin) Init(app *application.Application) error {
	cache := app.Cache()
	p.libraries = NewLibraries(app.Assets(), app.Logger(), cache.Data(), cache.Images(), p.config)
	return nil
}
