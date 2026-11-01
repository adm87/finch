package tiled

import (
	"errors"
	"fmt"

	"github.com/yohamta/donburi"

	"github.com/adm87/finch/pkg/application"
	"github.com/adm87/finch/pkg/assets"
	"github.com/adm87/finch/pkg/data"
	"github.com/adm87/finch/pkg/errs"
	"github.com/adm87/finch/pkg/images"
	"github.com/adm87/finch/pkg/logger"
	"github.com/adm87/finch/pkg/rendering"
	"github.com/adm87/finch/pkg/slotmap/store"
	"github.com/adm87/finch/pkg/transform"
	"github.com/adm87/finch/pkg/world"
	"github.com/adm87/finch/plugins/tiled/tmx"
)

// tiledTypes are stored as raw bytes in the data cache and parsed by CreateMap.
var tiledTypes = []string{"tmx", "tsx", "tx"}

const mapKind = "tiled map"

// Handle uniquely identifies a Tiled map.
type Handle uint64

// Plugin loads Tiled maps and draws them. It can't create maps until Init runs.
type Plugin struct {
	logger   logger.Logger
	assets   *assets.Assets
	data     *data.Cache
	images   *images.Cache
	config   *Config
	maps     *store.Storage[Handle, MapInfo]
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
	p.logger = app.Logger().Named("tiled")
	p.assets = app.Assets()
	p.data = cache.Data()
	p.images = cache.Images()
	p.maps = store.New[Handle, MapInfo](store.WithKind(mapKind), store.WithCapacity(32))

	rendererID, err := app.RenderPipeline().RegisterRenderer(NewTileMapRenderer(p))
	if err != nil {
		return fmt.Errorf("register renderer: %w", err)
	}
	p.renderer = rendererID

	return p.assets.RegisterImporter(data.NewImporter(p.data), tiledTypes)
}

func (p *Plugin) CreateMap(name string, mode assets.LoadMode) (Handle, error) {
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

	lease := p.assets.Lease(fsys, mode)
	l := &mapLoader{plugin: p, lease: lease, tilesets: make(map[assets.Asset]*tmx.Tileset)}
	info, err := l.load(cfg.TMX)
	if err != nil {
		return 0, fmt.Errorf("map %q: %w", name, errors.Join(err, lease.Release(assets.UnloadModeIfOwned)))
	}
	info.lease = lease

	return p.maps.Insert(name, info)
}

func (p *Plugin) ClearMap(name string, mode assets.UnloadMode) error {
	handle, exists := p.maps.Lookup(name)
	if !exists {
		return errs.NotFound(mapKind, name)
	}

	info, err := p.maps.Delete(handle)
	if err != nil {
		return fmt.Errorf("clear map %q: %w", name, err)
	}

	if err := info.lease.Release(mode); err != nil {
		return fmt.Errorf("release map %q: %w", name, err)
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

func (p *Plugin) AddToWorld(handle Handle, world *world.World) (*donburi.Entry, error) {
	info, err := p.Map(handle)
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
