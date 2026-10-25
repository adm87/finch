package application

import (
	"fmt"
	"io/fs"

	"github.com/adm87/finch/pkg/assets"
	"github.com/adm87/finch/pkg/display"
	"github.com/adm87/finch/pkg/errs"
	"github.com/adm87/finch/pkg/logger"
	"github.com/adm87/finch/pkg/rendering"
	"github.com/adm87/finch/pkg/world"
)

// Plugin hooks a third-party tool into the app. Init runs once, before startup.
type Plugin interface {
	Name() string
	Init(app *Application) error
}

type Application struct {
	logger         logger.Logger
	cache          *Cache
	renderers      *Renderers
	assets         *assets.Assets
	screen         *display.Screen
	view           *display.View
	renderPipeline *rendering.Pipeline
	world          *world.World
	config         *Config
	plugins        map[string]Plugin
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
		plugins:        make(map[string]Plugin),
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

// RegisterPlugin registers a plugin with the application.
// It returns an error if a plugin with the same name is already registered or if the plugin initialization fails.
func (a *Application) RegisterPlugin(plugin Plugin) error {
	if _, exists := a.plugins[plugin.Name()]; exists {
		return errs.AlreadyExists("plugin", plugin.Name())
	}
	if err := plugin.Init(a); err != nil {
		return err
	}
	a.plugins[plugin.Name()] = plugin
	return nil
}

// GetPlugin retrieves a registered plugin by name and asserts its type.
// It returns an error if the plugin is not found or if it is not of the expected type.
func GetPlugin[T Plugin](app *Application, name string) (T, error) {
	plugin, exists := app.plugins[name]
	if !exists {
		var zero T
		return zero, errs.NotFound("plugin", name)
	}
	p, ok := plugin.(T)
	if !ok {
		var zero T
		return zero, errs.InvalidOperation(fmt.Sprintf("plugin %q is not of the expected type", name))
	}
	return p, nil
}
