package internal

import (
	"fmt"
	"io/fs"
	"os"

	"github.com/adm87/finch/cmd/game/internal/camera"
	"github.com/adm87/finch/cmd/game/internal/config"
	"github.com/adm87/finch/cmd/game/internal/debug"
	"github.com/adm87/finch/cmd/game/internal/player"
	"github.com/adm87/finch/content"
	"github.com/adm87/finch/pkg/application"
	"github.com/adm87/finch/pkg/assets"
	"github.com/adm87/finch/pkg/engine"
	"github.com/adm87/finch/pkg/images"
	"github.com/adm87/finch/pkg/logger"
	"github.com/adm87/finch/pkg/rendering"
	"github.com/adm87/finch/pkg/world"
	"github.com/adm87/finch/plugins/aseprite"
	"github.com/adm87/finch/plugins/tiled"
)

const (
	playerLibrary = "captain"
	gym04         = "gym04"
)

func Boot() error {
	l := logger.NewLogger(os.Stdout, loggerOptions).Named("finch")
	if err := content.InitializeFileSystems(); err != nil {
		return fmt.Errorf("setup file systems: %w", err)
	}

	embedded, err := content.GetFileSystem(content.EmbeddedFS)
	if err != nil {
		return err
	}

	cfg := config.Default()
	if err := config.Load(embedded, content.GameConfig.String(), cfg); err != nil {
		return fmt.Errorf("load game config: %w", err)
	}

	app, err := newApplication(cfg, l)
	if err != nil {
		return err
	}

	if err := registerPlugins(app, embedded); err != nil {
		return fmt.Errorf("register plugins: %w", err)
	}

	return run(app, cfg, l)
}

func newApplication(cfg *config.GameConfig, l logger.Logger) (*application.Application, error) {
	app, err := application.New(&cfg.Application, l)
	if err != nil {
		return nil, fmt.Errorf("create application: %w", err)
	}
	if err := app.RegisterFileSystems(content.FileSystemRegistry); err != nil {
		return nil, fmt.Errorf("register file systems: %w", err)
	}
	return app, nil
}

// loadBundles runs at startup, after plugins have registered their importers.
func loadBundles(bundles []assets.AssetBundle) engine.Startup {
	return func(app *application.Application) error {
		for _, bundle := range bundles {
			if err := bundle.Load(app.Assets()); err != nil {
				return fmt.Errorf("load bundle from %q: %w", bundle.FileSystem, err)
			}
		}
		return nil
	}
}

// registerPlugins loads each plugin's config from fsys and registers it.
func registerPlugins(app *application.Application, fsys fs.FS) error {
	var aseCfg aseprite.Config
	if err := config.Load(fsys, content.AsepriteConfig.String(), &aseCfg); err != nil {
		return fmt.Errorf("load aseprite config: %w", err)
	}
	if err := app.RegisterPlugin(aseprite.NewPlugin(&aseCfg)); err != nil {
		return fmt.Errorf("register aseprite plugin: %w", err)
	}

	var tiledCfg tiled.Config
	if err := config.Load(fsys, content.TiledConfig.String(), &tiledCfg); err != nil {
		return fmt.Errorf("load tiled config: %w", err)
	}
	if err := app.RegisterPlugin(tiled.NewPlugin(&tiledCfg)); err != nil {
		return fmt.Errorf("register tiled plugin: %w", err)
	}
	return nil
}

func run(app *application.Application, cfg *config.GameConfig, l logger.Logger) error {
	app.World().Create(debug.DebugComponent, world.DontPartition)

	opts := append(gameSystems(app),
		engine.WithStartup(
			loadBundles(cfg.Bundles),
			makeTestScene,
		),
		engine.WithOverlayDrawers(
			debug.OverlayPartitions(app.World()),
			debug.OverlayEntityBounds(app.World()),
			debug.OverlayFPS(app.World()),
			debug.OverlayZeroZero(app.World()),
		),
		engine.WithLogger(l),
	)

	return engine.Run(app, &cfg.Engine, opts...)
}

func makeTestScene(app *application.Application) error {
	w := app.World()

	c := camera.NewMain(w)
	camera.SetZoom(c, 2)
	app.View().SetCamera(c)

	tld, err := application.GetPlugin[*tiled.Plugin](app, "tiled")
	if err != nil {
		return fmt.Errorf("get tiled plugin: %w", err)
	}
	maps := tld.Maps()

	mapHandle, err := maps.CreateMap(gym04, tiled.LoadModeDynamic)
	if err != nil {
		return fmt.Errorf("create map: %w", err)
	}
	if _, err = tld.AddToWorld(mapHandle, w); err != nil {
		return fmt.Errorf("add map to world: %w", err)
	}
	mapInfo, err := maps.Map(mapHandle)
	if err != nil {
		return fmt.Errorf("map %q: %w", gym04, err)
	}
	playerDepth, ok := mapInfo.LayerDepth("player")
	if !ok {
		return fmt.Errorf("map %q has no player layer", gym04)
	}

	ase, err := application.GetPlugin[*aseprite.Plugin](app, "aseprite")
	if err != nil {
		return fmt.Errorf("get aseprite plugin: %w", err)
	}
	libraries := ase.Libraries()

	libraryID, err := libraries.CreateLibrary(playerLibrary, aseprite.LoadModeDynamic)
	if err != nil {
		return fmt.Errorf("create library %q: %w", playerLibrary, err)
	}

	library, err := libraries.Library(libraryID)
	if err != nil {
		return fmt.Errorf("library %q: %w", playerLibrary, err)
	}

	frame, err := app.Cache().Images().Frame(library.ImageHandle, 0)
	if err != nil {
		return fmt.Errorf("first frame of %q: %w", playerLibrary, err)
	}

	width := float64(frame.Size.X)
	height := float64(frame.Size.Y)

	// Anchor at the character's feet in the first frame.
	visible := frame.Image.Bounds().Size()
	ax := (float64(frame.Offset.X) + float64(visible.X)/2) / width
	ay := float64(frame.Offset.Y+visible.Y) / height

	entry := w.Create(
		rendering.RendererComponent,
		images.ImageComponent,
		aseprite.AnimatorComponent,
		player.Player,
	)

	r := rendering.RendererComponent.Get(entry)
	r.SetRendererID(app.Renderers().Image)
	r.SetDepth(playerDepth)

	img := images.ImageComponent.Get(entry)
	img.SetHandle(library.ImageHandle)
	img.SetAnchor(ax, ay)

	anim := aseprite.AnimatorComponent.Get(entry)
	anim.SetLibrary(libraryID)
	anim.SetAnimation("Idle")

	return nil
}
