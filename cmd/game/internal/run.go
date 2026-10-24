package internal

import (
	"errors"
	"fmt"
	"os"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/yohamta/donburi"
	"gopkg.in/yaml.v3"

	"github.com/adm87/finch/cmd/game/internal/camera"
	"github.com/adm87/finch/cmd/game/internal/config"
	"github.com/adm87/finch/cmd/game/internal/game"
	"github.com/adm87/finch/cmd/game/internal/player"
	"github.com/adm87/finch/content"
	"github.com/adm87/finch/pkg/aseprite"
	"github.com/adm87/finch/pkg/assets"
	"github.com/adm87/finch/pkg/data"
	"github.com/adm87/finch/pkg/display"
	"github.com/adm87/finch/pkg/engine"
	"github.com/adm87/finch/pkg/errs"
	"github.com/adm87/finch/pkg/geom"
	"github.com/adm87/finch/pkg/images"
	"github.com/adm87/finch/pkg/logger"
	"github.com/adm87/finch/pkg/rendering"
	"github.com/adm87/finch/pkg/transform"
	"github.com/adm87/finch/pkg/world"
)

const resourceHandle = "resource handle"

func Boot() error {
	return boot(logger.NewLogger(os.Stdout, loggerOptions).Named("finch"))
}

func boot(l logger.Logger) error {
	var g game.Game

	if err := content.InitializeFileSystems(); err != nil {
		return fmt.Errorf("setup file systems: %w", err)
	}
	if err := initializeGameAssets(&g); err != nil {
		return fmt.Errorf("initialize game assets: %w", err)
	}
	if err := loadGameConfig(&g); err != nil {
		return fmt.Errorf("load game config: %w", err)
	}
	if err := initializeAseprite(&g, l); err != nil {
		return fmt.Errorf("initialize aseprite: %w", err)
	}

	return run(&g, l)
}

func initializeGameAssets(g *game.Game) error {
	g.Assets = assets.NewAssets()

	if err := registerFileSystem(g.Assets, content.EmbeddedFS); err != nil {
		return err
	}
	if err := registerFileSystem(g.Assets, content.AssetsFS); err != nil {
		return err
	}

	dc := data.NewCache()
	if err := g.Assets.RegisterImporter(data.NewImporter(dc), data.Types()); err != nil {
		return &errs.InvalidOperationError{Message: "register data importer", Inner: err}
	}

	ic := images.NewCache()
	if err := g.Assets.RegisterImporter(images.NewImporter(ic), images.Types()); err != nil {
		return &errs.InvalidOperationError{Message: "register image importer", Inner: err}
	}

	g.Cache = &game.Cache{
		Data:   dc,
		Images: ic,
	}

	return nil
}

func registerFileSystem(as *assets.Assets, id content.FileSystemID) error {
	fsys, err := content.GetFileSystem(id)
	if err != nil {
		return errs.NotFound("file system", id.String()).Wrap(err)
	}
	if err := as.RegisterFileSystem(assets.FileSystem(id), fsys); err != nil {
		return errs.InvalidOperation(fmt.Sprintf("register file system %q", id)).Wrap(err)
	}
	return nil
}

func loadGameConfig(g *game.Game) (err error) {
	fsys, err := content.GetFileSystem(content.EmbeddedFS)
	if err != nil {
		return errs.NotFound("file system", content.EmbeddedFS.String()).Wrap(err)
	}
	if err := g.Assets.Load(fsys, content.GameConfig); err != nil {
		return fmt.Errorf("load %q: %w", content.GameConfig, err)
	}

	defer func() {
		if uerr := g.Assets.Unload(content.GameConfig); uerr != nil {
			err = errors.Join(err, fmt.Errorf("unload %q: %w", content.GameConfig, uerr))
		}
	}()

	handle, err := g.Assets.GetHandle(content.GameConfig)
	if err != nil {
		return errs.NotFound(resourceHandle, content.GameConfig).Wrap(err)
	}

	raw, err := g.Cache.Data.Get(data.Handle(handle))
	if err != nil {
		return errs.NotFound("cached data", content.GameConfig).Wrap(err)
	}

	var cfg config.GameConfig
	if err := yaml.Unmarshal(raw, &cfg); err != nil {
		return errs.Format(content.GameConfig, "invalid yaml").Wrap(err)
	}
	if err := validateGameConfig(&cfg); err != nil {
		return err
	}

	if err := cfg.StaticContent.Load(g.Assets); err != nil {
		return fmt.Errorf("load static content: %w", err)
	}

	g.Config = &cfg
	return nil
}

func initializeAseprite(g *game.Game, logger logger.Logger) error {
	handle, err := g.Assets.GetHandle(content.AsepriteConfig)
	if err != nil {
		return errs.NotFound(resourceHandle, content.AsepriteConfig).Wrap(err)
	}

	raw, err := g.Cache.Data.Get(data.Handle(handle))
	if err != nil {
		return errs.NotFound("cached data", content.AsepriteConfig).Wrap(err)
	}

	var aseConfig aseprite.Config
	if err := yaml.Unmarshal(raw, &aseConfig); err != nil {
		return errs.Format(content.AsepriteConfig, "invalid yaml").Wrap(err)
	}

	ase, err := aseprite.New(g.Assets, logger, g.Cache.Data, g.Cache.Images, &aseConfig)
	if err != nil {
		return errs.InvalidOperation("initialize aseprite").Wrap(err)
	}

	g.Aseprite = ase
	return nil
}

func initializeGameRenderers(g *game.Game, rp *rendering.Pipeline) error {
	ir, err := rp.RegisterRenderer(images.NewRenderer(g.Cache.Images))
	if err != nil {
		return errs.InvalidOperation("register image renderer").Wrap(err)
	}

	g.Renderers = &game.Renderers{
		Image: ir,
	}

	return nil
}

func validateGameConfig(cfg *config.GameConfig) error {
	return errors.Join(
		errs.CheckRange("window.width", cfg.Window.Width, 1, 16384),
		errs.CheckRange("window.height", cfg.Window.Height, 1, 16384),
	)
}

func run(g *game.Game, l logger.Logger) error {
	w := world.NewWorld()

	g.Screen = display.NewScreen(
		g.Config.Window.Width,
		g.Config.Window.Height,
	)
	g.View = display.NewViewport(g.Screen)

	rp := rendering.NewPipeline(g.View, w)
	if err := initializeGameRenderers(g, rp); err != nil {
		return err
	}

	g.Time = game.NewTime(g.Config.FPS)

	gu := game.NewUpdater(g.Time)
	if err := initializeFrameSchedule(g, w, gu); err != nil {
		return err
	}

	if err := makeTestScene(g, w); err != nil {
		return fmt.Errorf("create test renderables: %w", err)
	}

	return engine.Run(
		engine.WithUpdate(gu),
		engine.WithDraw(
			game.NewDrawer(rp, g.Screen, g.View,
				game.WithClearColor(
					g.Config.Rendering.ClearColor,
				),
			),
		),
		engine.WithLayout(g.Screen),
		engine.WithWindowSize(
			g.Config.Window.Width,
			g.Config.Window.Height,
		),
		engine.WithResizeMode(
			ebiten.WindowResizingModeEnabled,
		),
		engine.WithFullscreen(
			g.Config.Window.Fullscreen,
		),
		engine.WithLogger(l),
	)
}

func makeTestScene(g *game.Game, w *world.World) error {
	c := w.Create(
		camera.MainCamera,
		world.DontPartition,
	)
	g.View.SetCamera(c)

	libraryID, err := g.Aseprite.CreateLibrary("captain", aseprite.LoadModeDynamic)
	if err != nil {
		return fmt.Errorf("create aseprite library: %w", err)
	}

	library, err := g.Aseprite.GetLibrary(libraryID)
	if err != nil {
		return fmt.Errorf("get aseprite image handle: %w", err)
	}

	frame, err := g.Cache.Images.GetFrame(library.ImageHandle, 0)
	if err != nil {
		return errs.NotFound("image", content.DefaultImage10x10).Wrap(err)
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

	r, rOk := rendering.GetRenderer(entry)
	i, iOk := images.GetImage(entry)
	b, bOk := transform.GetBounds(entry)
	a, aOk := aseprite.GetAnimator(entry)

	if !rOk {
		return missingComponent(entry.Entity(), "renderer")
	}
	if !iOk {
		return missingComponent(entry.Entity(), "image")
	}
	if !bOk {
		return missingComponent(entry.Entity(), "bounds")
	}
	if !aOk {
		return missingComponent(entry.Entity(), "aseprite animator")
	}

	r.SetID(g.Renderers.Image)

	i.SetHandle(library.ImageHandle)
	i.SetAnchor(ax, ay)

	b.SetRect(geom.Rect{
		X:      -width * ax,
		Y:      -height * ay,
		Width:  width,
		Height: height,
	})

	a.SetLibrary(libraryID)
	a.SetAnimation("Idle")

	if err := w.Add(entry); err != nil {
		return err
	}

	return nil
}

func missingComponent(entity donburi.Entity, component string) error {
	return errs.InvalidOperation(fmt.Sprintf("entity %v: missing %s component", entity, component))
}
