package tiled

import (
	"errors"
	"fmt"
	"strconv"

	"github.com/adm87/finch/pkg/assets"
	"github.com/adm87/finch/pkg/data"
	"github.com/adm87/finch/pkg/errs"
	"github.com/adm87/finch/pkg/images"
	"github.com/adm87/finch/pkg/logger"
	"github.com/adm87/finch/pkg/slotmap"
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
	logger     logger.Logger
	assets     *assets.Assets
	data       *data.Cache
	images     *images.Cache
	config     *Config
	maps       *slotmap.SlotMap[MapInfo]
	mapsByName map[string]Handle
}

func NewMaps(as *assets.Assets, data *data.Cache, images *images.Cache, cfg *Config, log logger.Logger) *Maps {
	return &Maps{
		logger:     log.Named("tiled"),
		assets:     as,
		data:       data,
		images:     images,
		config:     cfg,
		maps:       slotmap.New[MapInfo](32),
		mapsByName: make(map[string]Handle),
	}
}

func (m *Maps) CreateMap(name string, mode LoadMode) (Handle, error) {
	if handle, exists := m.mapsByName[name]; exists {
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

	handle := Handle(m.maps.Insert(info).Pack())
	m.mapsByName[name] = handle

	return handle, nil
}

func (m *Maps) ClearMap(name string, mode UnloadMode) error {
	handle, exists := m.mapsByName[name]
	if !exists {
		return errs.NotFound(mapKind, name)
	}

	info, err := m.maps.Delete(slotmap.Unpack(uint64(handle)))
	delete(m.mapsByName, name)
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
	info, err := m.maps.Get(slotmap.Unpack(uint64(handle)))
	if err != nil {
		return MapInfo{}, errs.NotFound(mapKind, strconv.FormatUint(uint64(handle), 10)).Wrap(err)
	}
	return info, nil
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
