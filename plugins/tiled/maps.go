package tiled

import (
	"errors"
	"fmt"

	"github.com/adm87/finch/pkg/assets"
	"github.com/adm87/finch/pkg/errs"
	"github.com/adm87/finch/plugins/tiled/tmx"
)

const mapKind = "tiled map"

// Handle uniquely identifies a Tiled map.
type Handle uint64

// Load builds the named map from the config, loading any of its files that
// aren't loaded yet, or returns it if it's already loaded.
func (p *Plugin) Load(name string) (Handle, error) {
	if handle, exists := p.maps.Lookup(name); exists {
		return handle, nil
	}

	cfg, err := p.getMapConfig(name)
	if err != nil {
		return 0, err
	}
	fsys, err := p.assets.FileSystem(cfg.FileSystem)
	if err != nil {
		return 0, fmt.Errorf("map %q: %w", name, err)
	}

	l := &mapLoader{
		plugin:    p,
		fsys:      fsys,
		tilesets:  make(map[string]*tmx.Tileset),
		templates: make(map[string]struct{}),
	}
	info, err := l.load(cfg.TMX.String())
	if err != nil {
		return 0, fmt.Errorf("map %q: %w", name, errors.Join(err, p.unload(l.owned)))
	}
	info.owned = l.owned

	return p.maps.Insert(name, info)
}

// Unload forgets the map and unloads the files its Load loaded.
func (p *Plugin) Unload(name string) error {
	handle, exists := p.maps.Lookup(name)
	if !exists {
		return errs.NotFound(mapKind, name)
	}

	info, err := p.maps.Delete(handle)
	if err != nil {
		return fmt.Errorf("unload map %q: %w", name, err)
	}

	if err := p.unload(info.owned); err != nil {
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

func (p *Plugin) unload(owned []assets.Asset) error {
	if len(owned) == 0 {
		return nil
	}
	return p.assets.Unload(owned...)
}
