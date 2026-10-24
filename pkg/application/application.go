package application

import (
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

// New uses DefaultConfig when config is nil.
func New(config *Config) (*Application, error) {
	cfg := *DefaultConfig()
	if config != nil {
		cfg = *config
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
	return a.config
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
