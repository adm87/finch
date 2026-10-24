package internal

import (
	"fmt"
	"io/fs"
	"os"

	"github.com/hajimehoshi/ebiten/v2"
	"gopkg.in/yaml.v3"

	"github.com/adm87/finch/cmd/game/internal/camera"
	"github.com/adm87/finch/cmd/game/internal/config"
	"github.com/adm87/finch/cmd/game/internal/game"
	"github.com/adm87/finch/cmd/game/internal/player"
	"github.com/adm87/finch/content"
	"github.com/adm87/finch/pkg/application"
	"github.com/adm87/finch/pkg/aseprite"
	"github.com/adm87/finch/pkg/assets"
	"github.com/adm87/finch/pkg/data"
	"github.com/adm87/finch/pkg/engine"
	"github.com/adm87/finch/pkg/errs"
	"github.com/adm87/finch/pkg/geom"
	"github.com/adm87/finch/pkg/images"
	"github.com/adm87/finch/pkg/logger"
	"github.com/adm87/finch/pkg/rendering"
	"github.com/adm87/finch/pkg/transform"
	"github.com/adm87/finch/pkg/world"
)

const playerLibrary = "captain"

func Boot() error {
	if err := content.InitializeFileSystems(); err != nil {
		return fmt.Errorf("setup file systems: %w", err)
	}
	return boot(logger.NewLogger(os.Stdout, loggerOptions).Named("finch"))
}

func boot(l logger.Logger) error {
	var g game.Game

	cfg, err := loadGameConfig()
	if err != nil {
		return fmt.Errorf("load game config: %w", err)
	}

	fileSystems, err := gameFileSystems()
	if err != nil {
		return err
	}

	app, err := application.New(&cfg.Application, fileSystems)
	if err != nil {
		return fmt.Errorf("create application: %w", err)
	}

	if err := initializeAseprite(&g, app, l); err != nil {
		return fmt.Errorf("initialize aseprite: %w", err)
	}

	return run(&g, app, l)
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

// loadGameConfig skips the asset system because the config sizes the application that owns it.
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

func initializeAseprite(g *game.Game, app *application.Application, logger logger.Logger) error {
	cache := app.Cache()

	handle, err := app.Assets().Handle(content.AsepriteConfig)
	if err != nil {
		return err
	}
	raw, err := cache.Data().Get(data.Handle(handle))
	if err != nil {
		return err
	}

	var aseConfig aseprite.Config
	if err := yaml.Unmarshal(raw, &aseConfig); err != nil {
		return errs.Format(content.AsepriteConfig, "invalid yaml").Wrap(err)
	}

	ase, err := aseprite.New(app.Assets(), logger, cache.Data(), cache.Images(), &aseConfig)
	if err != nil {
		return err
	}

	g.Aseprite = ase
	return nil
}

func run(g *game.Game, app *application.Application, l logger.Logger) error {
	if err := makeTestScene(g, app); err != nil {
		return fmt.Errorf("create test scene: %w", err)
	}

	opts := append(gameSystems(g),
		engine.WithOverlay(engine.DrawerFunc(game.DrawDebugOverlay)),
		engine.WithResizeMode(ebiten.WindowResizingModeEnabled),
		engine.WithLogger(l),
	)
	return engine.Run(app, opts...)
}

func makeTestScene(g *game.Game, app *application.Application) error {
	w := app.World()

	c := w.Create(
		camera.MainCamera,
		world.DontPartition,
	)
	app.View().SetCamera(c)

	libraryID, err := g.Aseprite.CreateLibrary(playerLibrary, aseprite.LoadModeDynamic)
	if err != nil {
		return fmt.Errorf("create library %q: %w", playerLibrary, err)
	}

	library, err := g.Aseprite.Library(libraryID)
	if err != nil {
		return fmt.Errorf("library %q: %w", playerLibrary, err)
	}

	frame, err := app.Cache().Images().Frame(library.ImageHandle, 0)
	if err != nil {
		return fmt.Errorf("first frame of %q: %w", playerLibrary, err)
	}

	width := float64(frame.Size.X)
	height := float64(frame.Size.Y)

	// Anchor at the feet: the bottom center of the first frame's visible pixels.
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

	transform.BoundsComponent.Get(entry).SetRect(geom.Rect{
		X:      -width * ax,
		Y:      -height * ay,
		Width:  width,
		Height: height,
	})

	anim := aseprite.AnimatorComponent.Get(entry)
	anim.SetLibrary(libraryID)
	anim.SetAnimation("Idle")

	return w.Add(entry)
}
