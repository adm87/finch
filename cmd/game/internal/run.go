package internal

import (
	"errors"
	"image/color"
	"math/rand/v2"
	"os"

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
	"github.com/hajimehoshi/ebiten/v2"
)

const (
	windowWidth  = 1280
	windowHeight = 720
)

func Run() error {
	w := world.NewWorld()

	l := logger.NewLogger(
		os.Stdout,
	)
	s := game.NewScreen(
		windowWidth,
		windowHeight,
	)

	v := game.NewViewport(s)

	rp := rendering.NewRenderPipeline(v, w)
	res := resources.NewResources()

	ic := images.NewImageCache()
	il := images.NewImageLoader(ic)
	ir := rp.AddAdapter(images.NewImageRenderer(ic))
	res.RegisterImporter(il, images.ImageTypes())

	c := w.Create(world.DontPartition)
	v.SetCamera(c)

	if err := loadTestContent(res); err != nil {
		return err
	}
	if err := makeTestRenderables(w, res, ic, ir); err != nil {
		return err
	}

	return engine.Run(
		engine.WithUpdate(
			game.NewGameUpdate(c),
		),
		engine.WithDraw(
			game.NewGameDraw(rp, s, v),
		),
		engine.WithLayout(s),
		engine.WithWindowSize(
			windowWidth,
			windowHeight,
		),
		engine.WithResizeMode(
			ebiten.WindowResizingModeEnabled,
		),
		engine.WithLogger(l),
	)
}

func loadTestContent(res resources.Resources) error {
	if err := res.Load(content.EmbeddedFiles, content.DefaultImage10x10); err != nil {
		return err
	}
	return nil
}

func makeTestRenderables(w world.World, res resources.Resources, ic *images.ImageCache, rendererID uint64) error {
	imgHandle, err := res.GetHandle(content.DefaultImage10x10)
	if err != nil {
		return err
	}

	img, _ := ic.Get(imgHandle)
	width := img.Bounds().Dx()
	height := img.Bounds().Dy()

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
