package internal

import (
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/yohamta/donburi"

	"github.com/adm87/finch/cmd/game/internal/camera"
	"github.com/adm87/finch/cmd/game/internal/debug"
	"github.com/adm87/finch/cmd/game/internal/input"
	"github.com/adm87/finch/cmd/game/internal/player"
	"github.com/adm87/finch/pkg/application"
	"github.com/adm87/finch/pkg/engine"
	"github.com/adm87/finch/plugins/aseprite"
)

func gameSystems(app *application.Application) []engine.RunOption {
	var entries []*donburi.Entry

	ase, err := application.GetPlugin[*aseprite.Plugin](app, "aseprite")
	if err != nil {
		panic(err)
	}

	handleInput := func(app *application.Application, dt float64) error {
		w := app.World()
		if inpututil.IsKeyJustPressed(ebiten.KeyEscape) {
			return ebiten.Termination
		}

		if c, ok := camera.MainCamera.First(w.ECS()); ok {
			input.UpdateCamera(dt, c)
		}
		if p, ok := player.Player.First(w.ECS()); ok {
			input.UpdatePlayer(dt, p)
		}

		debug.PollInput(w)
		return nil
	}

	collectVisible := func(app *application.Application, _ float64) error {
		rect, _ := app.View().Viewport()
		entries = app.World().Query(rect, entries[:0])
		return nil
	}

	animate := func(_ *application.Application, dt float64) error {
		return aseprite.AdvanceAnimations(dt, ase.Libraries(), entries)
	}

	followPlayer := func(app *application.Application, dt float64) error {
		w := app.World()
		c, cOk := camera.MainCamera.First(w.ECS())
		p, pOk := player.Player.First(w.ECS())
		if cOk && pOk {
			camera.FollowTarget(dt, c, p)
		}
		return nil
	}

	return []engine.RunOption{
		engine.WithUpdate(handleInput),
		engine.WithLateUpdate(collectVisible, animate, followPlayer),
	}
}
