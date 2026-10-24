package application

import (
	"fmt"
	"io/fs"

	"github.com/adm87/finch/pkg/assets"
	"github.com/adm87/finch/pkg/data"
	"github.com/adm87/finch/pkg/display"
	"github.com/adm87/finch/pkg/images"
)

type Application struct {
	config Config
	cache  *Cache
	assets *assets.Assets
	screen *display.Screen
	view   *display.View
}

// New uses DefaultConfig when config is nil. It registers fileSystems with
// the asset system and then loads the config's bundles.
func New(config *Config, fileSystems map[assets.FileSystem]fs.FS) (*Application, error) {
	cfg := *DefaultConfig()
	if config != nil {
		cfg = config.clone()
	}
	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	as := assets.NewAssets()
	cache := NewCache()

	if err := as.RegisterImporter(data.NewImporter(cache.data), data.Types()); err != nil {
		return nil, err
	}
	if err := as.RegisterImporter(images.NewImporter(cache.images), images.Types()); err != nil {
		return nil, err
	}
	for name, fsys := range fileSystems {
		if err := as.RegisterFileSystem(name, fsys); err != nil {
			return nil, err
		}
	}
	for _, bundle := range cfg.Bundles {
		if err := bundle.Load(as); err != nil {
			return nil, fmt.Errorf("load bundle from %q: %w", bundle.FileSystem, err)
		}
	}

	screen := display.NewScreen(cfg.Screen.Width, cfg.Screen.Height)

	return &Application{
		config: cfg,
		cache:  cache,
		assets: as,
		screen: screen,
		view:   display.NewViewport(screen),
	}, nil
}

func (a *Application) Config() Config {
	return a.config.clone()
}

func (a *Application) Cache() *Cache {
	return a.cache
}

func (a *Application) Assets() *assets.Assets {
	return a.assets
}

func (a *Application) Screen() *display.Screen {
	return a.screen
}

func (a *Application) View() *display.View {
	return a.view
}
