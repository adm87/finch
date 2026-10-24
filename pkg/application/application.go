package application

import (
	"fmt"
	"io/fs"

	"github.com/adm87/finch/pkg/assets"
	"github.com/adm87/finch/pkg/data"
	"github.com/adm87/finch/pkg/display"
	"github.com/adm87/finch/pkg/images"
	"github.com/adm87/finch/pkg/rendering"
	"github.com/adm87/finch/pkg/world"
)

type Application struct {
	cache          *Cache
	renderers      *Renderers
	assets         *assets.Assets
	screen         *display.Screen
	view           *display.View
	renderPipeline *rendering.Pipeline
	world          *world.World
	config         *Config
}

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
	view := display.NewViewport(screen)

	world := world.NewWorld()
	renderPipeline := rendering.NewPipeline(view, world)

	renderers, err := newRenderers(renderPipeline, cache)
	if err != nil {
		return nil, err
	}

	return &Application{
		config:         &cfg,
		cache:          cache,
		assets:         as,
		screen:         screen,
		renderers:      renderers,
		renderPipeline: renderPipeline,
		view:           view,
		world:          world,
	}, nil
}

func (a *Application) Config() Config {
	return a.config.clone()
}

func (a *Application) Cache() *Cache {
	return a.cache
}

func (a *Application) Renderers() *Renderers {
	return a.renderers
}

func (a *Application) Assets() *assets.Assets {
	return a.assets
}

func (a *Application) RenderPipeline() *rendering.Pipeline {
	return a.renderPipeline
}

func (a *Application) Screen() *display.Screen {
	return a.screen
}

func (a *Application) View() *display.View {
	return a.view
}

func (a *Application) World() *world.World {
	return a.world
}
