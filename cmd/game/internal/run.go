package internal

import (
	"os"

	"github.com/adm87/finch/cmd/game/internal/game"
	"github.com/adm87/finch/pkg/ecs/components/camera"
	"github.com/adm87/finch/pkg/ecs/components/transform"
	"github.com/adm87/finch/pkg/ecs/rendering"
	"github.com/adm87/finch/pkg/engine"
	"github.com/adm87/finch/pkg/logger"
	"github.com/yohamta/donburi"
)

const (
	window_width  = 800
	window_height = 600
)

func Run() error {
	w := donburi.NewWorld()

	l := logger.NewLogger(
		os.Stdout,
	)
	s := game.NewScreen(
		window_width,
		window_height,
	)

	v := game.NewViewport(s)
	rp := rendering.NewRenderPipeline(v)

	c := makeCamera(w)
	v.SetCamera(c)

	gu := game.NewGameUpdate(c)
	gd := game.NewGameDraw(rp, s, v)

	return engine.Run(
		engine.WithUpdate(gu),
		engine.WithDraw(gd),
		engine.WithLayout(s),
		engine.WithWindowSize(
			window_width,
			window_height,
		),
		engine.WithLogger(l),
	)
}

func makeCamera(world donburi.World) *donburi.Entry {
	entry := world.Entry(world.Create(
		transform.BoundsComponent,
		transform.MatrixComponent,
		transform.TransformComponent,
		camera.MainCamera,
	))

	t := transform.TransformComponent.Get(entry)
	t.SetOrigin(window_width/2, window_height/2)

	return entry
}
