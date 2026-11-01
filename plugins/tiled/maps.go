package tiled

import (
	"fmt"

	"github.com/adm87/finch/pkg/errs"
	"github.com/adm87/finch/pkg/resource"
	"github.com/adm87/finch/plugins/tiled/tmx"
)

const mapKind = "tiled map"

// Handle uniquely identifies a Tiled map.
type Handle uint64

// Load builds the named map from src, or returns it if it's already loaded
// from there.
func (p *Plugin) Load(src resource.Source, name string) (Handle, error) {
	key := resource.Key(src, name)
	if handle, exists := p.maps.Lookup(key); exists {
		return handle, nil
	}

	cfg, err := p.getMapConfig(name)
	if err != nil {
		return 0, err
	}
	fsys, err := p.sources.FS(src)
	if err != nil {
		return 0, fmt.Errorf("map %q: %w", name, err)
	}

	l := &mapLoader{plugin: p, src: src, fsys: fsys, tilesets: make(map[string]*tmx.Tileset)}
	info, err := l.load(cfg.TMX)
	if err != nil {
		return 0, fmt.Errorf("map %q: %w", name, err)
	}
	return p.maps.Insert(key, info)
}

// Unload forgets the map loaded from src. Its tile images stay cached for
// other maps.
func (p *Plugin) Unload(src resource.Source, name string) error {
	handle, exists := p.maps.Lookup(resource.Key(src, name))
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
