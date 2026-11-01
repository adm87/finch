package game

import (
	"fmt"
	"io/fs"

	"github.com/yohamta/donburi"

	"github.com/adm87/finch/cmd/game/internal/config"
	"github.com/adm87/finch/cmd/game/internal/debug"
	"github.com/adm87/finch/content"
	"github.com/adm87/finch/pkg/assets"
	"github.com/adm87/finch/pkg/data"
	"github.com/adm87/finch/pkg/display"
	"github.com/adm87/finch/pkg/images"
	"github.com/adm87/finch/pkg/logger"
	"github.com/adm87/finch/pkg/rendering"
	"github.com/adm87/finch/pkg/world"
	"github.com/adm87/finch/plugins/aseprite"
	"github.com/adm87/finch/plugins/tiled"
)

type Game struct {
	world         *world.World
	screen        *display.Screen
	view          *display.View
	pipeline      *rendering.Pipeline
	assets        *assets.Assets
	images        *images.Cache
	tiled         *tiled.Plugin
	libs          *aseprite.Libraries
	visible       []*donburi.Entry
	imageRenderer rendering.AdapterID
}

// New builds the engine pieces and the plugins, loading their configs from fsys.
func New(cfg *config.GameConfig, fsys fs.FS, l logger.Logger) (*Game, error) {
	if err := cfg.Screen.Validate(); err != nil {
		return nil, err
	}

	screen := display.NewScreen(cfg.Screen.Width, cfg.Screen.Height)
	g := &Game{
		world:    world.NewWorld(),
		screen:   screen,
		view:     display.NewViewport(screen),
		pipeline: rendering.NewPipeline(),
		assets:   assets.NewAssets(),
		images:   images.NewCache(),
	}

	for name, files := range content.FileSystemRegistry {
		if err := g.assets.RegisterFileSystem(name, files); err != nil {
			return nil, fmt.Errorf("register file system %q: %w", name, err)
		}
	}

	dataCache := data.NewCache()
	if err := g.assets.RegisterImporter(data.NewImporter(dataCache), data.Types()); err != nil {
		return nil, err
	}
	if err := g.assets.RegisterImporter(images.NewImporter(g.images), images.Types()); err != nil {
		return nil, err
	}

	imageRenderer, err := g.pipeline.RegisterRenderer(images.NewRenderer(g.images))
	if err != nil {
		return nil, fmt.Errorf("register image renderer: %w", err)
	}
	g.imageRenderer = imageRenderer

	var tiledCfg tiled.Config
	if err := config.Load(fsys, content.TiledConfig.String(), &tiledCfg); err != nil {
		return nil, fmt.Errorf("load tiled config: %w", err)
	}
	if g.tiled, err = tiled.NewPlugin(&tiledCfg, g.assets, dataCache, g.images, g.pipeline, l); err != nil {
		return nil, fmt.Errorf("create tiled plugin: %w", err)
	}

	var aseCfg aseprite.Config
	if err := config.Load(fsys, content.AsepriteConfig.String(), &aseCfg); err != nil {
		return nil, fmt.Errorf("load aseprite config: %w", err)
	}
	g.libs = aseprite.NewLibraries(g.assets, l, dataCache, g.images, &aseCfg)

	g.world.Create(debug.DebugComponent, world.DontPartition)

	return g, nil
}
