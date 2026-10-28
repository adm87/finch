package tiled

import (
	"github.com/adm87/finch/pkg/application"
	"github.com/adm87/finch/pkg/data"
	"github.com/adm87/finch/pkg/errs"
)

// tiledTypes are stored as raw bytes in the data cache and parsed by Maps.
var tiledTypes = []string{"tmx", "tsx", "tx"}

type Plugin struct {
	config *Config
	maps   *Maps
}

func NewPlugin(cfg *Config) *Plugin {
	return &Plugin{config: cfg}
}

func (p *Plugin) Name() string {
	return "tiled"
}

func (p *Plugin) Init(app *application.Application) error {
	if p.config == nil {
		return errs.ArgumentNil("config")
	}
	if err := p.config.Validate(); err != nil {
		return err
	}

	cache := app.Cache()
	assets := app.Assets()

	p.maps = NewMaps(assets, cache.Data(), cache.Images(), p.config, app.Logger())
	return assets.RegisterImporter(data.NewImporter(cache.Data()), tiledTypes)
}

func (p *Plugin) Maps() *Maps {
	return p.maps
}
