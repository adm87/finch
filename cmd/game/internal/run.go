package internal

import (
	"errors"
	"fmt"
	"image/color"
	"math/rand/v2"
	"os"

	"github.com/hajimehoshi/ebiten/v2"
	"gopkg.in/yaml.v3"

	"github.com/adm87/finch/cmd/game/internal/config"
	"github.com/adm87/finch/cmd/game/internal/game"
	"github.com/adm87/finch/content"
	"github.com/adm87/finch/pkg/ecs/components/renderer"
	"github.com/adm87/finch/pkg/ecs/components/transform"
	"github.com/adm87/finch/pkg/ecs/rendering"
	"github.com/adm87/finch/pkg/ecs/world"
	"github.com/adm87/finch/pkg/engine"
	"github.com/adm87/finch/pkg/geom"
	"github.com/adm87/finch/pkg/images"
	"github.com/adm87/finch/pkg/logger"
	"github.com/adm87/finch/pkg/resources"
	"github.com/adm87/finch/pkg/resources/data"
	"github.com/adm87/finch/pkg/types/errs"
)

func Boot() error {
	var g game.Game

	if err := content.InitializeFileSystems(); err != nil {
		return fmt.Errorf("setup file systems: %w", err)
	}
	if err := initializeGameResources(&g); err != nil {
		return fmt.Errorf("setup game resources: %w", err)
	}
	if err := loadGameConfig(&g); err != nil {
		return fmt.Errorf("load game config: %w", err)
	}
	if err := g.Config.StaticResources.Load(g.Resources); err != nil {
		return fmt.Errorf("load static resources: %w", err)
	}

	return run(&g)
}

func initializeGameResources(g *game.Game) error {
	g.Resources = resources.NewResources()

	if err := registerFileSystem(g.Resources, content.FS_EMBEDDED); err != nil {
		return err
	}
	if err := registerFileSystem(g.Resources, content.FS_ASSETS); err != nil {
		return err
	}

	dc := data.NewDataCache()
	if err := g.Resources.RegisterImporter(data.NewDataLoader(dc), data.DataTypes()); err != nil {
		return &errs.InvalidOperationError{Message: "register data importer", Inner: err}
	}

	ic := images.NewImageCache()
	if err := g.Resources.RegisterImporter(images.NewImageLoader(ic), images.ImageTypes()); err != nil {
		return &errs.InvalidOperationError{Message: "register image importer", Inner: err}
	}

	g.Cache = &game.GameCache{
		Data:   dc,
		Images: ic,
	}

	return nil
}

func registerFileSystem(r resources.Resources, id content.FileSystemID) error {
	fsys, err := content.GetFileSystem(id)
	if err != nil {
		return &errs.NotFoundError{Kind: "file system", Key: id.String(), Inner: err}
	}
	if err := r.RegisterFileSystem(resources.FileSystem(id), fsys); err != nil {
		return &errs.InvalidOperationError{Message: fmt.Sprintf("register file system %q", id), Inner: err}
	}
	return nil
}

func initializeGameRenderers(g *game.Game, rp *rendering.RenderPipeline) error {
	ir, err := rp.RegisterRenderer(images.NewImageRenderer(g.Cache.Images))
	if err != nil {
		return &errs.InvalidOperationError{Message: "register image renderer", Inner: err}
	}

	g.Renderers = &game.GameRenderers{
		Image: ir,
	}

	return nil
}

func loadGameConfig(g *game.Game) (err error) {
	fsys, err := content.GetFileSystem(content.FS_EMBEDDED)
	if err != nil {
		return &errs.NotFoundError{Kind: "file system", Key: content.FS_EMBEDDED.String(), Inner: err}
	}
	if err := g.Resources.Load(fsys, content.GAME_CONFIG); err != nil {
		return fmt.Errorf("load %q: %w", content.GAME_CONFIG, err)
	}

	defer func() {
		if uerr := g.Resources.Unload(content.GAME_CONFIG); uerr != nil {
			err = errors.Join(err, fmt.Errorf("unload %q: %w", content.GAME_CONFIG, uerr))
		}
	}()

	handle, err := g.Resources.GetHandle(content.GAME_CONFIG)
	if err != nil {
		return &errs.NotFoundError{Kind: "resource handle", Key: content.GAME_CONFIG, Inner: err}
	}

	raw, err := g.Cache.Data.Get(handle)
	if err != nil {
		return &errs.NotFoundError{Kind: "cached data", Key: content.GAME_CONFIG, Inner: err}
	}

	var cfg config.GameConfig
	if err := yaml.Unmarshal(raw, &cfg); err != nil {
		return &errs.FormatError{Input: content.GAME_CONFIG, Message: "invalid yaml", Inner: err}
	}
	if err := validateGameConfig(&cfg); err != nil {
		return err
	}

	g.Config = &cfg
	return nil
}

func validateGameConfig(cfg *config.GameConfig) error {
	return errors.Join(
		errs.CheckRange("window.width", cfg.Window.Width, 1, 16384),
		errs.CheckRange("window.height", cfg.Window.Height, 1, 16384),
	)
}

func run(g *game.Game) error {
	l := logger.NewLogger(os.Stdout)
	w := world.NewWorld()

	s := game.NewScreen(
		g.Config.Window.Width,
		g.Config.Window.Height,
	)

	v := game.NewViewport(s)
	rp := rendering.NewRenderPipeline(v, w)

	if err := initializeGameRenderers(g, rp); err != nil {
		return err
	}

	c := w.Create(world.DontPartition)
	v.SetCamera(c)

	if err := makeTestRenderables(w, g); err != nil {
		return fmt.Errorf("create test renderables: %w", err)
	}

	return engine.Run(
		engine.WithUpdate(
			game.NewGameUpdate(c),
		),
		engine.WithDraw(
			game.NewGameDraw(rp, s, v,
				game.WithClearColor(
					g.Config.Rendering.ClearColor,
				),
			),
		),
		engine.WithLayout(s),
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

func makeTestRenderables(w world.World, g *game.Game) error {

	imgHandle, err := g.Resources.GetHandle(content.DEFAULT_IMAGE_10X10)
	if err != nil {
		return &errs.NotFoundError{Kind: "resource handle", Key: content.DEFAULT_IMAGE_10X10, Inner: err}
	}

	img, err := g.Cache.Images.Get(imgHandle)
	if err != nil {
		return &errs.NotFoundError{Kind: "image", Key: content.DEFAULT_IMAGE_10X10, Inner: err}
	}

	width := float64(img.Bounds().Dx())
	height := float64(img.Bounds().Dy())

	windowWidth := float64(g.Config.Window.Width)
	windowHeight := float64(g.Config.Window.Height)

	const ax, ay = 0.5, 0.5

	var errList []error
	for n := range 1000 {
		entry := w.Create(
			renderer.RendererComponent,
			images.ImageComponent,
		)

		r, ok := renderer.GetRenderer(entry)
		if !ok {
			errList = append(errList, missingComponent(n, "renderer"))
			continue
		}
		t, ok := transform.GetTransform(entry)
		if !ok {
			errList = append(errList, missingComponent(n, "transform"))
			continue
		}
		i, ok := images.GetImage(entry)
		if !ok {
			errList = append(errList, missingComponent(n, "image"))
			continue
		}

		r.SetID(g.Renderers.Image)

		t.SetPosition(
			-windowWidth*0.5+rand.Float64()*windowWidth,
			-windowHeight*0.5+rand.Float64()*windowHeight,
		)

		i.SetHandle(imgHandle)
		i.SetAnchor(ax, ay)
		i.SetColor(color.RGBA{
			R: uint8(rand.UintN(256)),
			G: uint8(rand.UintN(256)),
			B: uint8(rand.UintN(256)),
			A: 255,
		})

		transform.SetRect(entry, geom.Rect{
			X:      -width * ax,
			Y:      -height * ay,
			Width:  width,
			Height: height,
		})

		if err := w.Add(entry); err != nil {
			errList = append(errList, fmt.Errorf("entity %d: add to world: %w", n, err))
		}
	}

	return errors.Join(errList...)
}

func missingComponent(entity int, component string) error {
	return errs.InvalidOperation(fmt.Sprintf("entity %d: missing %s component", entity, component))
}
