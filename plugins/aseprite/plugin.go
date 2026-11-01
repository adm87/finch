package aseprite

import (
	"errors"
	"fmt"

	"github.com/adm87/finch/pkg/application"
	"github.com/adm87/finch/pkg/assets"
	"github.com/adm87/finch/pkg/data"
	"github.com/adm87/finch/pkg/errs"
	"github.com/adm87/finch/pkg/images"
	"github.com/adm87/finch/pkg/logger"
	"github.com/adm87/finch/pkg/slotmap/store"
)

const libraryKind = "aseprite library"

// Handle uniquely identifies an Aseprite library.
type Handle uint64

// Plugin loads Aseprite libraries. It can't create libraries until Init runs.
type Plugin struct {
	logger    logger.Logger
	assets    *assets.Assets
	data      *data.Cache
	images    *images.Cache
	config    *Config
	libraries *store.Storage[Handle, LibraryInfo]
}

func NewPlugin(cfg *Config) *Plugin {
	return &Plugin{config: cfg}
}

func (p *Plugin) Name() string {
	return "aseprite"
}

func (p *Plugin) Init(app *application.Application) error {
	if p.config == nil {
		return errs.ArgumentNil("config")
	}
	if err := p.config.Validate(); err != nil {
		return err
	}

	cache := app.Cache()
	p.logger = app.Logger().Named("aseprite")
	p.assets = app.Assets()
	p.data = cache.Data()
	p.images = cache.Images()
	p.libraries = store.New[Handle, LibraryInfo](store.WithKind(libraryKind))
	return nil
}

func (p *Plugin) CreateLibrary(name string, mode assets.LoadMode) (Handle, error) {
	if handle, exists := p.libraries.Lookup(name); exists {
		return handle, nil
	}

	cfg, err := p.getLibraryConfig(name)
	if err != nil {
		return 0, err
	}
	fsys, err := p.assets.FileSystem(cfg.FileSystem)
	if err != nil {
		return 0, fmt.Errorf("library %q: %w", name, err)
	}

	lease := p.assets.Lease(fsys, mode)
	ll := &libraryLoader{plugin: p, lease: lease}
	info, err := ll.load(cfg)
	if err != nil {
		return 0, fmt.Errorf("library %q: %w", name, errors.Join(err, lease.Release(assets.UnloadModeIfOwned)))
	}
	info.lease = lease

	return p.libraries.Insert(name, info)
}

func (p *Plugin) ClearLibrary(name string, mode assets.UnloadMode) error {
	handle, exists := p.libraries.Lookup(name)
	if !exists {
		return errs.NotFound(libraryKind, name)
	}

	info, err := p.libraries.Delete(handle)
	if err != nil {
		return fmt.Errorf("clear library %q: %w", name, err)
	}

	if err := info.lease.Release(mode); err != nil {
		return fmt.Errorf("release library %q: %w", name, err)
	}
	return nil
}

func (p *Plugin) Library(handle Handle) (LibraryInfo, error) {
	return p.libraries.Get(handle)
}

func (p *Plugin) getLibraryConfig(name string) (*LibraryConfig, error) {
	if name == "" {
		return nil, errs.Argument("name", "cannot be empty")
	}
	cfg, exists := p.config.Libraries[name]
	if !exists {
		return nil, errs.NotFound(libraryKind, name)
	}
	return &cfg, nil
}
