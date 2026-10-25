package internal

import (
	"fmt"
	"io/fs"
	"os"

	"github.com/hajimehoshi/ebiten/v2"
	"gopkg.in/yaml.v3"

	"github.com/adm87/finch/cmd/game/internal/camera"
	"github.com/adm87/finch/cmd/game/internal/config"
	"github.com/adm87/finch/cmd/game/internal/debug"
	"github.com/adm87/finch/cmd/game/internal/player"
	"github.com/adm87/finch/content"
	"github.com/adm87/finch/pkg/application"
	"github.com/adm87/finch/pkg/assets"
	"github.com/adm87/finch/pkg/data"
	"github.com/adm87/finch/pkg/engine"
	"github.com/adm87/finch/pkg/errs"
	"github.com/adm87/finch/pkg/images"
	"github.com/adm87/finch/pkg/logger"
	"github.com/adm87/finch/pkg/rendering"
	"github.com/adm87/finch/pkg/world"
	"github.com/adm87/finch/plugins/aseprite"
	"github.com/adm87/finch/plugins/tiled"
)

const playerLibrary = "captain"

func Boot() error {
	l := logger.NewLogger(os.Stdout, loggerOptions).Named("finch")
	if err := content.InitializeFileSystems(); err != nil {
		return fmt.Errorf("setup file systems: %w", err)
	}

	cfg, err := loadGameConfig()
	if err != nil {
		return fmt.Errorf("load game config: %w", err)
	}

	app, err := newApplication(cfg, l)
	if err != nil {
		return err
	}

	if err := registerPlugins(app); err != nil {
		return fmt.Errorf("register plugins: %w", err)
	}

	return run(app, cfg, l)
}

// loadGameConfig reads the file directly. The asset system lives in the app, which isn't built yet.
func loadGameConfig() (*config.GameConfig, error) {
	fsys, err := content.GetFileSystem(content.EmbeddedFS)
	if err != nil {
		return nil, err
	}

	raw, err := fs.ReadFile(fsys, content.GameConfig)
	if err != nil {
		return nil, fmt.Errorf("read %q: %w", content.GameConfig, err)
	}

	cfg, err := config.Parse(raw)
	if err != nil {
		return nil, errs.Format(content.GameConfig, "invalid yaml").Wrap(err)
	}
	return cfg, nil
}

func newApplication(cfg *config.GameConfig, l logger.Logger) (*application.Application, error) {
	fileSystems, err := gameFileSystems()
	if err != nil {
		return nil, fmt.Errorf("initialize game file systems: %w", err)
	}

	app, err := application.New(&cfg.Application, l)
	if err != nil {
		return nil, fmt.Errorf("create application: %w", err)
	}
	if err := app.RegisterFileSystems(fileSystems); err != nil {
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

func gameFileSystems() (map[assets.FileSystem]fs.FS, error) {
	fileSystems := make(map[assets.FileSystem]fs.FS)
	for _, id := range []content.FileSystemID{content.EmbeddedFS, content.AssetsFS} {
		fsys, err := content.GetFileSystem(id)
		if err != nil {
			return nil, err
		}
		fileSystems[assets.FileSystem(id)] = fsys
	}
	return fileSystems, nil
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
		engine.WithResizeMode(
			ebiten.WindowResizingModeEnabled,
		),
		engine.WithLogger(l),
	)

	return engine.Run(app, opts...)
}

func registerPlugins(app *application.Application) error {
	aseCfg, err := loadConfig[aseprite.Config](app, assets.FileSystem(content.EmbeddedFS), content.AsepriteConfig)
	if err != nil {
		return fmt.Errorf("load aseprite config: %w", err)
	}
	if err := app.RegisterPlugin(aseprite.NewPlugin(aseCfg)); err != nil {
		return fmt.Errorf("register aseprite plugin: %w", err)
	}

	tiledCfg, err := loadConfig[tiled.Config](app, assets.FileSystem(content.EmbeddedFS), content.TiledConfig)
	if err != nil {
		return fmt.Errorf("load tiled config: %w", err)
	}
	if err := app.RegisterPlugin(tiled.NewPlugin(tiledCfg)); err != nil {
		return fmt.Errorf("register tiled plugin: %w", err)
	}
	return nil
}

func loadConfig[T any](app *application.Application, filesystem assets.FileSystem, config assets.Asset) (*T, error) {
	as := app.Assets()
	cache := app.Cache()

	bundle := assets.AssetBundle{FileSystem: filesystem, Assets: []assets.Asset{config}}
	if err := bundle.Load(as); err != nil && !errs.IsAlreadyExists(err) {
		return nil, err
	}

	handle, err := as.Handle(config)
	if err != nil {
		return nil, err
	}
	raw, err := cache.Data().Get(data.Handle(handle))
	if err != nil {
		return nil, err
	}

	var cfg T
	if err := yaml.Unmarshal(raw, &cfg); err != nil {
		return nil, errs.Format(config.String(), "invalid yaml").Wrap(err)
	}

	return &cfg, nil
}

func makeTestScene(app *application.Application) error {
	w := app.World()

	c := w.Create(
		camera.MainCamera,
		world.DontPartition,
	)
	app.View().SetCamera(c)

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

	rendering.RendererComponent.Get(entry).SetID(app.Renderers().Image)

	img := images.ImageComponent.Get(entry)
	img.SetHandle(library.ImageHandle)
	img.SetAnchor(ax, ay)

	anim := aseprite.AnimatorComponent.Get(entry)
	anim.SetLibrary(libraryID)
	anim.SetAnimation("Idle")

	return nil
}
