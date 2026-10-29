package internal

import (
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/yohamta/donburi"
	"github.com/yohamta/donburi/filter"

	"github.com/adm87/finch/cmd/game/internal/camera"
	"github.com/adm87/finch/cmd/game/internal/input"
	"github.com/adm87/finch/cmd/game/internal/player"
	"github.com/adm87/finch/pkg/application"
	"github.com/adm87/finch/pkg/engine"
	"github.com/adm87/finch/pkg/geom"
	"github.com/adm87/finch/pkg/transform"
	"github.com/adm87/finch/pkg/world"
	"github.com/adm87/finch/plugins/aseprite"
	"github.com/adm87/finch/plugins/tiled"
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

		if c, ok := camera.Main(w); ok {
			input.UpdateCamera(dt, c)
		}
		if p, ok := player.Player.First(w.ECS()); ok {
			input.UpdatePlayer(dt, p)
			if bounds, ok := mapBounds(w); ok {
				t := transform.TransformComponent.Get(p)
				t.SetPosition(bounds.ClampPoint(t.Position()))
			}
		}
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
		c, cOk := camera.Main(w)
		p, pOk := player.Player.First(w.ECS())
		if cOk && pOk {
			camera.Follow(dt, c, p)
		}
		if bounds, ok := mapBounds(w); cOk && ok {
			view, _ := app.View().Viewport()
			camera.Clamp(c, view, bounds)
		}
		return nil
	}

	return []engine.RunOption{
		engine.WithUpdate(handleInput),
		engine.WithLateUpdate(followPlayer, collectVisible, animate),
	}
}

var tileMaps = donburi.NewQuery(filter.Contains(tiled.TileMapComponent))

// mapBounds is the world rect of the first tile map.
func mapBounds(w *world.World) (geom.Rect, bool) {
	tm, ok := tileMaps.First(w.ECS())
	if !ok {
		return geom.Rect{}, false
	}
	b, ok := transform.GetBounds(tm)
	if !ok {
		return geom.Rect{}, false
	}
	return b.WorldRect(transform.TransformComponent.Get(tm)), true
}
