package tiled

import (
	"fmt"

	"github.com/yohamta/donburi"

	"github.com/adm87/finch/pkg/application"
	"github.com/adm87/finch/pkg/data"
	"github.com/adm87/finch/pkg/errs"
	"github.com/adm87/finch/pkg/rendering"
	"github.com/adm87/finch/pkg/transform"
	"github.com/adm87/finch/pkg/world"
)

// tiledTypes are stored as raw bytes in the data cache and parsed by Maps.
var tiledTypes = []string{"tmx", "tsx", "tx"}

type Plugin struct {
	config   *Config
	maps     *Maps
	renderer rendering.AdapterID
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
	pipeline := app.RenderPipeline()

	p.maps = NewMaps(assets, cache.Data(), cache.Images(), p.config, app.Logger())

	rendererID, err := pipeline.RegisterRenderer(NewTileMapRenderer(p.maps, cache.Images()))
	if err != nil {
		return fmt.Errorf("register renderer: %w", err)
	}
	p.renderer = rendererID

	return assets.RegisterImporter(data.NewImporter(cache.Data()), tiledTypes)
}

func (p *Plugin) Maps() *Maps {
	return p.maps
}

func (p *Plugin) AddToWorld(handle Handle, world *world.World) (*donburi.Entry, error) {
	info, err := p.maps.Map(handle)
	if err != nil {
		return nil, err
	}

	e := world.Create(TileMapComponent, rendering.RendererComponent)

	b, _ := transform.GetBounds(e)
	b.SetRect(info.Bounds())

	r, _ := rendering.GetRenderer(e)
	r.SetRendererID(p.renderer)

	tm, _ := GetTileMap(e)
	tm.SetHandle(handle)

	return e, world.Update(e)
}
