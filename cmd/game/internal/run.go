package internal

import (
	"math/rand/v2"
	"os"

	"github.com/adm87/finch/cmd/game/internal/game"
	"github.com/adm87/finch/pkg/ecs/components/renderer"
	"github.com/adm87/finch/pkg/ecs/components/transform"
	"github.com/adm87/finch/pkg/ecs/rendering"
	"github.com/adm87/finch/pkg/ecs/world"
	"github.com/adm87/finch/pkg/engine"
	"github.com/adm87/finch/pkg/geom"
	"github.com/adm87/finch/pkg/images"
	"github.com/adm87/finch/pkg/logger"
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

	ir := rp.AddAdapter(images.NewImageRenderer())

	c := w.Create(world.DontPartition)
	v.SetCamera(c)

	makeTestRenderables(w, ir)

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

func makeTestRenderables(w world.World, rendererID uint64) {
	for range 1000 {
		x := -window_width*0.5 + rand.Float64()*window_width
		y := -window_height*0.5 + rand.Float64()*window_height

		entry := w.Create(renderer.RendererComponent)

		r, _ := renderer.GetRenderer(entry)
		r.SetID(rendererID)

		t, _ := transform.GetTransform(entry)
		t.SetPosition(float64(x), float64(y))

		transform.SetRect(entry, geom.Rect{
			X:      0,
			Y:      0,
			Width:  10,
			Height: 10,
		})

		w.Add(entry)
	}
}
