package internal

import (
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/yohamta/donburi"

	"github.com/adm87/finch/cmd/game/internal/camera"
	"github.com/adm87/finch/cmd/game/internal/game"
	"github.com/adm87/finch/cmd/game/internal/input"
	"github.com/adm87/finch/cmd/game/internal/player"
	"github.com/adm87/finch/pkg/application"
	"github.com/adm87/finch/pkg/aseprite"
	"github.com/adm87/finch/pkg/engine"
	"github.com/adm87/finch/pkg/world"
)

func gameSystems(g *game.Game, w *world.World) []engine.RunOption {
	var entries []*donburi.Entry

	handleInput := func(_ *application.Application, dt float64) error {
		if inpututil.IsKeyJustPressed(ebiten.KeyEscape) {
			return ebiten.Termination
		}
		if c, ok := camera.MainCamera.First(w.ECS()); ok {
			input.UpdateCamera(dt, c)
		}
		if p, ok := player.Player.First(w.ECS()); ok {
			return input.UpdatePlayer(dt, p, w)
		}
		return nil
	}

	collectVisible := func(app *application.Application, _ float64) error {
		rect, _ := app.View().Viewport()
		entries = w.Query(rect, entries[:0])
		return nil
	}

	animate := func(_ *application.Application, dt float64) error {
		return aseprite.AdvanceAnimations(dt, g.Aseprite, entries)
	}

	followPlayer := func(_ *application.Application, dt float64) error {
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
