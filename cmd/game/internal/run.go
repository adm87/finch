package internal

import (
	"image/color"
	"os"

	"github.com/adm87/finch/cmd/game/internal/game"
	"github.com/adm87/finch/pkg/ecs/rendering"
	"github.com/adm87/finch/pkg/ecs/world"
	"github.com/adm87/finch/pkg/engine"
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
	rp := rendering.NewRenderPipeline(v)

	c := w.Create(world.DontPartition)
	v.SetCamera(c)

	gu := game.NewGameUpdate(c)
	gd := game.NewGameDraw(rp, s, v,
		game.WithClearColor(color.RGBA{R: 100, G: 149, B: 237, A: 255}),
	)

	return engine.Run(
		engine.WithUpdate(gu),
		engine.WithDraw(gd),
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
