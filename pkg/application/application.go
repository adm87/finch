package application

import (
	"io/fs"

	"github.com/adm87/finch/pkg/assets"
	"github.com/adm87/finch/pkg/display"
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

	if err := registerAssetImporters(as, cache); err != nil {
		return nil, err
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
	return *a.config
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

// RegisterFileSystems stops at the first error; file systems before it stay registered.
func (a *Application) RegisterFileSystems(fileSystems map[assets.FileSystem]fs.FS) error {
	for name, fsys := range fileSystems {
		if err := a.assets.RegisterFileSystem(name, fsys); err != nil {
			return err
		}
	}
	return nil
}
