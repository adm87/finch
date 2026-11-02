package tiled

import (
	"fmt"

	"github.com/adm87/finch/pkg/errs"
	"github.com/adm87/finch/plugins/tiled/tmx"
)

const mapKind = "tiled map"

// Handle uniquely identifies a Tiled map.
type Handle uint64

// Load builds the named map from the config, or returns it if it's already
// built. Its .tmx, its tilesets and their images must already be loaded.
func (p *Plugin) Load(name string) (Handle, error) {
	if handle, exists := p.maps.Lookup(name); exists {
		return handle, nil
	}

	cfg, err := p.getMapConfig(name)
	if err != nil {
		return 0, err
	}

	l := &mapLoader{plugin: p, tilesets: make(map[string]*tmx.Tileset)}
	info, err := l.load(cfg.TMX.String())
	if err != nil {
		return 0, fmt.Errorf("map %q: %w", name, err)
	}
	return p.maps.Insert(name, info)
}

// Unload forgets the map. Its files stay loaded.
func (p *Plugin) Unload(name string) error {
	handle, exists := p.maps.Lookup(name)
	if !exists {
		return errs.NotFound(mapKind, name)
	}
	if _, err := p.maps.Delete(handle); err != nil {
		return fmt.Errorf("unload map %q: %w", name, err)
	}
	return nil
}

func (p *Plugin) Map(handle Handle) (MapInfo, error) {
	return p.maps.Get(handle)
}

func (p *Plugin) getMapConfig(name string) (*MapConfig, error) {
	if name == "" {
		return nil, errs.Argument("name", "cannot be empty")
	}
	cfg, exists := p.config.Maps[name]
	if !exists {
		return nil, errs.NotFound(mapKind, name)
	}
	return &cfg, nil
}
