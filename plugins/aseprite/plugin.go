package aseprite

import (
	"github.com/adm87/finch/pkg/application"
)

// Plugin loads its config from fileSystem when it initializes.
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
	as := app.Assets()
	cache := app.Cache()

	libraries, err := NewLibraries(as, app.Logger(), cache.Data(), cache.Images(), p.config)
	if err != nil {
		return err
	}

	p.libraries = libraries
	return nil
}
