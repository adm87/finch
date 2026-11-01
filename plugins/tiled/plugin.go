package tiled

import (
	"fmt"

	"github.com/yohamta/donburi"

	"github.com/adm87/finch/pkg/assets"
	"github.com/adm87/finch/pkg/data"
	"github.com/adm87/finch/pkg/errs"
	"github.com/adm87/finch/pkg/images"
	"github.com/adm87/finch/pkg/logger"
	"github.com/adm87/finch/pkg/rendering"
	"github.com/adm87/finch/pkg/transform"
	"github.com/adm87/finch/pkg/world"
)

// tiledTypes are stored as raw bytes in the data cache and parsed by Maps.
var tiledTypes = []string{"tmx", "tsx", "tx"}

type Plugin struct {
	maps     *Maps
	renderer rendering.AdapterID
}

// NewPlugin registers the Tiled importer and renderer.
func NewPlugin(cfg *Config, as *assets.Assets, dataCache *data.Cache, imageCache *images.Cache,
	pipeline *rendering.Pipeline, log logger.Logger,
) (*Plugin, error) {
	if cfg == nil {
		return nil, errs.ArgumentNil("config")
	}
	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	maps := NewMaps(as, dataCache, imageCache, cfg, log)

	renderer, err := pipeline.RegisterRenderer(NewTileMapRenderer(maps, imageCache))
	if err != nil {
		return nil, fmt.Errorf("register renderer: %w", err)
	}
	if err := as.RegisterImporter(data.NewImporter(dataCache), tiledTypes); err != nil {
		return nil, err
	}
	return &Plugin{maps: maps, renderer: renderer}, nil
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
