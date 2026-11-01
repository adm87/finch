package game

import (
	"fmt"

	"github.com/yohamta/donburi"

	"github.com/adm87/finch/cmd/game/internal/debug"
	"github.com/adm87/finch/pkg/display"
	"github.com/adm87/finch/pkg/images"
	"github.com/adm87/finch/pkg/logger"
	"github.com/adm87/finch/pkg/rendering"
	"github.com/adm87/finch/pkg/resource"
	"github.com/adm87/finch/pkg/world"
	"github.com/adm87/finch/plugins/aseprite"
	"github.com/adm87/finch/plugins/tiled"
)

type Game struct {
	world         *world.World
	screen        *display.Screen
	view          *display.View
	pipeline      *rendering.Pipeline
	sources       *resource.Loader
	images        *images.Cache
	bundles       map[string]Bundle
	tiled         *tiled.Plugin
	aseprite      *aseprite.Plugin
	visible       []*donburi.Entry
	imageRenderer rendering.AdapterID
}

// Configs are the parsed config files the game is built from.
type Configs struct {
	Game     *Config
	Tiled    *tiled.Config
	Aseprite *aseprite.Config
}

// New builds the engine pieces and the plugins. Bundles are loaded from sources.
func New(cfgs Configs, sources *resource.Loader, l logger.Logger) (*Game, error) {
	cfg := cfgs.Game
	if err := cfg.Screen.Validate(); err != nil {
		return nil, err
	}

	screen := display.NewScreen(cfg.Screen.Width, cfg.Screen.Height)
	g := &Game{
		world:    world.NewWorld(),
		screen:   screen,
		view:     display.NewViewport(screen),
		pipeline: rendering.NewPipeline(),
		sources:  sources,
		images:   images.NewCache(),
		bundles:  cfg.Bundles,
	}

	imageRenderer, err := g.pipeline.RegisterRenderer(images.NewRenderer(g.images))
	if err != nil {
		return nil, fmt.Errorf("register image renderer: %w", err)
	}
	g.imageRenderer = imageRenderer

	if g.tiled, err = tiled.NewPlugin(cfgs.Tiled, sources, g.pipeline, l); err != nil {
		return nil, fmt.Errorf("create tiled plugin: %w", err)
	}

	if g.aseprite, err = aseprite.NewPlugin(cfgs.Aseprite, sources, g.images, l); err != nil {
		return nil, fmt.Errorf("create aseprite plugin: %w", err)
	}

	g.world.Create(debug.DebugComponent, world.DontPartition)

	return g, nil
}
