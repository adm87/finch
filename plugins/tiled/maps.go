package tiled

import (
	"errors"
	"fmt"

	"github.com/adm87/finch/pkg/assets"
	"github.com/adm87/finch/pkg/data"
	"github.com/adm87/finch/pkg/errs"
	"github.com/adm87/finch/pkg/images"
	"github.com/adm87/finch/pkg/logger"
	"github.com/adm87/finch/pkg/slotmap/store"
	"github.com/adm87/finch/plugins/tiled/tmx"
)

const mapKind = "tiled map"

// Handle uniquely identifies a Tiled map.
type Handle uint64

// LoadMode controls whether CreateMap may load missing files.
type LoadMode uint8

const (
	// LoadModeRequired fails if any of the map's files isn't already loaded.
	LoadModeRequired LoadMode = iota
	// LoadModeDynamic loads the map's files that aren't loaded.
	LoadModeDynamic
)

// UnloadMode controls whether ClearMap unloads the map's files.
type UnloadMode int

const (
	// UnloadModeKeep leaves the files loaded.
	UnloadModeKeep UnloadMode = iota
	// UnloadModeIfOwned unloads the files CreateMap loaded.
	UnloadModeIfOwned
)

type Maps struct {
	logger logger.Logger
	assets *assets.Assets
	data   *data.Cache
	images *images.Cache
	config *Config
	maps   *store.Storage[Handle, MapInfo]
}

func NewMaps(as *assets.Assets, data *data.Cache, images *images.Cache, cfg *Config, log logger.Logger) *Maps {
	return &Maps{
		logger: log.Named("tiled"),
		assets: as,
		data:   data,
		images: images,
		config: cfg,
		maps:   store.New[Handle, MapInfo](store.WithKind(mapKind), store.WithCapacity(32)),
	}
}

func (m *Maps) CreateMap(name string, mode LoadMode) (Handle, error) {
	if handle, exists := m.maps.Lookup(name); exists {
		return handle, nil
	}

	cfg, err := m.getMapConfig(name)
	if err != nil {
		return 0, err
	}
	fsys, err := m.assets.FileSystem(cfg.FileSystem)
	if err != nil {
		return 0, fmt.Errorf("map %q: %w", name, err)
	}

	l := &mapLoader{maps: m, fsys: fsys, mode: mode, tilesets: make(map[assets.Asset]*tmx.Tileset)}
	info, err := l.load(cfg.TMX)
	if err != nil {
		return 0, fmt.Errorf("map %q: %w", name, errors.Join(err, m.unload(l.owned)))
	}
	info.owned = l.owned

	return m.maps.Insert(name, info)
}

func (m *Maps) ClearMap(name string, mode UnloadMode) error {
	handle, exists := m.maps.Lookup(name)
	if !exists {
		return errs.NotFound(mapKind, name)
	}

	info, err := m.maps.Delete(handle)
	if err != nil {
		return fmt.Errorf("clear map %q: %w", name, err)
	}

	if mode != UnloadModeIfOwned {
		return nil
	}
	if err := m.unload(info.owned); err != nil {
		return fmt.Errorf("unload map %q: %w", name, err)
	}
	return nil
}

func (m *Maps) Map(handle Handle) (MapInfo, error) {
	return m.maps.Get(handle)
}

func (m *Maps) getMapConfig(name string) (*MapConfig, error) {
	if name == "" {
		return nil, errs.Argument("name", "cannot be empty")
	}
	cfg, exists := m.config.Maps[name]
	if !exists {
		return nil, errs.NotFound(mapKind, name)
	}
	return &cfg, nil
}

func (m *Maps) unload(owned []assets.Asset) error {
	if len(owned) == 0 {
		return nil
	}
	return m.assets.Unload(owned...)
}
