package game

import (
	"fmt"
	"io/fs"

	"github.com/yohamta/donburi"

	"github.com/adm87/finch/cmd/game/internal/debug"
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

// Files are the file systems the game loads assets from. Their names are
// what configs use for "filesystem".
type Files struct {
	Images   fs.FS
	Aseprite fs.FS
	Tiled    fs.FS
}

// New builds the engine pieces and the plugins.
func New(cfgs Configs, files Files, l logger.Logger) (*Game, error) {
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
		assets:   assets.NewAssets(),
		images:   images.NewCache(),
	}

	for name, fsys := range map[assets.FileSystem]fs.FS{
		"images":   files.Images,
		"aseprite": files.Aseprite,
		"tiled":    files.Tiled,
	} {
		if err := g.assets.RegisterFileSystem(name, fsys); err != nil {
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

	if g.tiled, err = tiled.NewPlugin(cfgs.Tiled, g.assets, dataCache, g.images, g.pipeline, l); err != nil {
		return nil, fmt.Errorf("create tiled plugin: %w", err)
	}

	if g.aseprite, err = aseprite.NewPlugin(cfgs.Aseprite, g.assets, dataCache, g.images, l); err != nil {
		return nil, fmt.Errorf("create aseprite plugin: %w", err)
	}

	g.world.Create(debug.DebugComponent, world.DontPartition)

	return g, nil
}
