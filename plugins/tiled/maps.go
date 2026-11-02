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

// LoadMode controls whether Load may load missing files.
type LoadMode uint8

const (
	// LoadModeRequired fails if any of the map's files isn't already loaded.
	LoadModeRequired LoadMode = iota
	// LoadModeDynamic loads the map's files that aren't loaded.
	LoadModeDynamic
)

// UnloadMode controls whether Unload unloads the map's files.
type UnloadMode int

const (
	// UnloadModeKeep leaves the files loaded.
	UnloadModeKeep UnloadMode = iota
	// UnloadModeIfOwned unloads the files Load loaded.
	UnloadModeIfOwned
)

// Load builds the named map from the config, or returns it if it's already loaded.
func (p *Plugin) Load(name string, mode LoadMode) (Handle, error) {
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

	l := &mapLoader{plugin: p, fsys: fsys, mode: mode, tilesets: make(map[assets.Asset]*tmx.Tileset)}
	info, err := l.load(cfg.TMX)
	if err != nil {
		return 0, fmt.Errorf("map %q: %w", name, errors.Join(err, p.unload(l.owned)))
	}
	info.owned = l.owned

	return p.maps.Insert(name, info)
}

func (p *Plugin) Unload(name string, mode UnloadMode) error {
	handle, exists := p.maps.Lookup(name)
	if !exists {
		return errs.NotFound(mapKind, name)
	}

	info, err := p.maps.Delete(handle)
	if err != nil {
		return fmt.Errorf("unload map %q: %w", name, err)
	}

	if mode != UnloadModeIfOwned {
		return nil
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
