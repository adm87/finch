package internal

import (
	"errors"
	"image/color"
	"log"
	"math/rand/v2"
	"os"
	"path/filepath"

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
	"github.com/alexflint/go-arg"
	"github.com/hajimehoshi/ebiten/v2"
)

type CmdArgs struct {
	WorkingDir string `arg:"--working-dir"`
}

func Run() error {
	var args CmdArgs
	arg.MustParse(&args)

	root, err := filepath.Abs(args.WorkingDir)
	if err != nil {
		log.Fatalf("failed to get absolute path for root: %v", err)
	}

	contentRoot := filepath.Join(root, "content")
	if _, err := os.Stat(contentRoot); os.IsNotExist(err) {
		log.Fatalf("content directory does not exist: %v", contentRoot)
	}

	var gc config.GameConfig
	config.MustLoadConfig(contentRoot, &gc)

	return run(&gc, contentRoot)
}

func run(gc *config.GameConfig, contentRoot string) error {
	l := logger.NewLogger(os.Stdout)
	w := world.NewWorld()

	gr, err := initializeGameResources(contentRoot)
	if err != nil {
		return err
	}

	if err := gc.Resources.Load(gr.Resources); err != nil {
		return err
	}

	s := game.NewScreen(
		gc.Window.Width,
		gc.Window.Height,
	)

	v := game.NewViewport(s)
	rp := rendering.NewRenderPipeline(v, w)
	grr := initializeGameRenderers(rp, gr)

	c := w.Create(world.DontPartition)
	v.SetCamera(c)

	if err := makeTestRenderables(w, gc, gr, grr); err != nil {
		return err
	}

	return engine.Run(
		engine.WithUpdate(
			game.NewGameUpdate(c),
		),
		engine.WithDraw(
			game.NewGameDraw(rp, s, v,
				game.WithClearColor(gc.Rendering.ClearColor),
			),
		),
		engine.WithLayout(s),
		engine.WithWindowSize(
			gc.Window.Width,
			gc.Window.Height,
		),
		engine.WithResizeMode(
			ebiten.WindowResizingModeEnabled,
		),
		engine.WithFullscreen(
			gc.Window.Fullscreen,
		),
		engine.WithLogger(l),
	)
}

func initializeGameResources(contentRoot string) (*game.GameResources, error) {
	res := resources.NewResources()

	res.AddFileSystem("embedded", content.EmbeddedFiles)
	res.AddFileSystem("assets", os.DirFS(filepath.Join(contentRoot, "assets")))

	dc := data.NewDataCache()
	dl := data.NewDataLoader(dc)
	res.RegisterImporter(dl, data.DataTypes())

	ic := images.NewImageCache()
	il := images.NewImageLoader(ic)
	res.RegisterImporter(il, images.ImageTypes())

	return &game.GameResources{
		Resources: res,
		Cache: &game.GameCache{
			Data:   dc,
			Images: ic,
		},
	}, nil
}

func initializeGameRenderers(rp *rendering.RenderPipeline, gr *game.GameResources) *game.GameRenderers {
	return &game.GameRenderers{
		Images: rp.AddAdapter(images.NewImageRenderer(gr.Cache.Images)),
	}
}

func makeTestRenderables(w world.World, cfg *config.GameConfig, gr *game.GameResources, grr *game.GameRenderers) error {
	imgHandle, err := gr.Resources.GetHandle(content.DefaultImage10x10)
	if err != nil {
		return err
	}

	img, _ := gr.Cache.Images.Get(imgHandle)
	width := img.Bounds().Dx()
	height := img.Bounds().Dy()

	windowWidth := float64(cfg.Window.Width)
	windowHeight := float64(cfg.Window.Height)

	var errs []error
	for range 1000 {
		x := -windowWidth*0.5 + rand.Float64()*windowWidth
		y := -windowHeight*0.5 + rand.Float64()*windowHeight

		entry := w.Create(
			renderer.RendererComponent,
			images.ImageComponent,
		)

		r, _ := renderer.GetRenderer(entry)
		r.SetID(grr.Images)

		t, _ := transform.GetTransform(entry)
		t.SetPosition(float64(x), float64(y))

		ax, ay := 0.5, 0.5

		i, _ := images.GetImage(entry)
		i.SetHandle(imgHandle)
		i.SetAnchor(ax, ay)
		i.SetColor(color.RGBA{
			R: uint8(rand.Uint() * 255),
			G: uint8(rand.Uint() * 255),
			B: uint8(rand.Uint() * 255),
			A: 255,
		})

		transform.SetRect(entry, geom.Rect{
			X:      -float64(width) * ax,
			Y:      -float64(height) * ay,
			Width:  float64(width),
			Height: float64(height),
		})

		if err := w.Add(entry); err != nil {
			errs = append(errs, err)
		}
	}

	return errors.Join(errs...)
}
