package internal

import (
	"errors"
	"image/color"
	"math/rand/v2"
	"os"

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
	"github.com/hajimehoshi/ebiten/v2"
)

var initialResourceBundle = []resources.Resource{
	content.GameConfig,
	content.DefaultImage10x10,
}

func Run() error {
	w := world.NewWorld()
	res := resources.NewResources()

	dc := data.NewDataCache()
	dl := data.NewDataLoader(dc)
	res.RegisterImporter(dl, data.DataTypes())

	ic := images.NewImageCache()
	il := images.NewImageLoader(ic)
	res.RegisterImporter(il, images.ImageTypes())

	if err := res.Load(content.EmbeddedFiles, initialResourceBundle...); err != nil {
		return err
	}

	cfg, err := makeGameConfig(res, dc)
	if err != nil {
		return err
	}

	l := logger.NewLogger(
		os.Stdout,
	)
	s := game.NewScreen(
		cfg.Window.Width,
		cfg.Window.Height,
	)

	v := game.NewViewport(s)
	rp := rendering.NewRenderPipeline(v, w)

	ir := rp.AddAdapter(images.NewImageRenderer(ic))

	c := w.Create(world.DontPartition)
	v.SetCamera(c)

	if err := makeTestRenderables(cfg, w, res, ic, ir); err != nil {
		return err
	}

	return engine.Run(
		engine.WithUpdate(
			game.NewGameUpdate(c),
		),
		engine.WithDraw(
			game.NewGameDraw(rp, s, v,
				game.WithClearColor(cfg.Rendering.ClearColor),
			),
		),
		engine.WithLayout(s),
		engine.WithWindowSize(
			cfg.Window.Width,
			cfg.Window.Height,
		),
		engine.WithResizeMode(
			ebiten.WindowResizingModeEnabled,
		),
		engine.WithFullscreen(
			cfg.Window.Fullscreen,
		),
		engine.WithLogger(l),
	)
}

func makeGameConfig(res resources.Resources, dc *data.DataCache) (*config.GameConfig, error) {
	handle, err := res.GetHandle(content.GameConfig)
	if err != nil {
		return nil, err
	}
	raw, err := dc.Get(handle)
	if err != nil {
		return nil, err
	}
	return config.LoadConfig(raw)
}

func makeTestRenderables(cfg *config.GameConfig, w world.World, res resources.Resources, ic *images.ImageCache, rendererID uint64) error {
	imgHandle, err := res.GetHandle(content.DefaultImage10x10)
	if err != nil {
		return err
	}

	img, _ := ic.Get(imgHandle)
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
		r.SetID(rendererID)

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
