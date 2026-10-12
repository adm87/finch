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
	window_width  = 1280
	window_height = 720
)

func Run() error {
	w := world.NewWorld()

	l := logger.NewLogger(
		os.Stdout,
	)
	s := game.NewScreen(
		window_width,
		window_height,
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
	if err := makeTestRenderables(w, res, ir); err != nil {
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
			window_width,
			window_height,
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

func makeTestRenderables(w world.World, res resources.Resources, rendererID uint64) error {
	imgHandle, err := res.GetHandle(content.DefaultImage10x10)
	if err != nil {
		return err
	}

	var errs []error
	for range 100 {
		x := -window_width*0.5 + rand.Float64()*window_width
		y := -window_height*0.5 + rand.Float64()*window_height

		entry := w.Create(
			renderer.RendererComponent,
			images.ImageComponent,
		)

		r, _ := renderer.GetRenderer(entry)
		r.SetID(rendererID)

		t, _ := transform.GetTransform(entry)
		t.SetPosition(float64(x), float64(y))

		i, _ := images.GetImage(entry)
		i.SetHandle(imgHandle)
		i.SetAnchor(0.5, 0.5)
		i.SetColor(color.RGBA{
			R: uint8(rand.Uint() * 255),
			G: uint8(rand.Uint() * 255),
			B: uint8(rand.Uint() * 255),
			A: 255,
		})

		transform.SetRect(entry, geom.Rect{
			X:      0,
			Y:      0,
			Width:  10,
			Height: 10,
		})

		if err := w.Add(entry); err != nil {
			errs = append(errs, err)
		}
	}

	return errors.Join(errs...)
}
