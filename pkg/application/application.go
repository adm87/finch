package application

import (
	"io/fs"

	"github.com/adm87/finch/pkg/assets"
	"github.com/adm87/finch/pkg/display"
	"github.com/adm87/finch/pkg/logger"
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
	logger         logger.Logger
}

// New falls back to DefaultConfig and a no-op logger when given nil.
func New(config *Config, log logger.Logger) (*Application, error) {
	cfg := *DefaultConfig()
	if config != nil {
		cfg = *config
	}
	if err := cfg.Validate(); err != nil {
		return nil, err
	}
	if log == nil {
		log = logger.N
	}

	as := assets.NewAssets()
	cache := NewCache()

	if err := registerAssetImporters(as, cache); err != nil {
		return nil, err
	}

	screen := display.NewScreen(cfg.Screen.Width, cfg.Screen.Height)
	view := display.NewViewport(screen)

	world := world.NewWorld()
	renderPipeline := rendering.NewPipeline()

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
		logger:         log,
	}, nil
}

func (a *Application) Config() Config {
	return *a.config
}

func (a *Application) Logger() logger.Logger {
	return a.logger
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

// RegisterFileSystems stops at the first error. Anything registered before it stays.
func (a *Application) RegisterFileSystems(fileSystems map[assets.FileSystem]fs.FS) error {
	for name, fsys := range fileSystems {
		if err := a.assets.RegisterFileSystem(name, fsys); err != nil {
			return err
		}
	}
	return nil
}
