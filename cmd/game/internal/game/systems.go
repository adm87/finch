package game

import (
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/yohamta/donburi"
	"github.com/yohamta/donburi/filter"

	"github.com/adm87/finch/cmd/game/internal/camera"
	"github.com/adm87/finch/cmd/game/internal/input"
	"github.com/adm87/finch/cmd/game/internal/player"
	"github.com/adm87/finch/pkg/geom"
	"github.com/adm87/finch/pkg/transform"
	"github.com/adm87/finch/pkg/world"
	"github.com/adm87/finch/plugins/aseprite"
	"github.com/adm87/finch/plugins/tiled"
)

func (g *Game) handleInput(dt float64) error {
	if inpututil.IsKeyJustPressed(ebiten.KeyEscape) {
		return ebiten.Termination
	}
	if c, ok := camera.Main(g.world); ok {
		input.UpdateCamera(dt, c)
	}
	if p, ok := player.Player.First(g.world.ECS()); ok {
		input.UpdatePlayer(dt, p)
		if bounds, ok := mapBounds(g.world); ok {
			t := transform.TransformComponent.Get(p)
			t.SetPosition(bounds.ClampPoint(t.Position()))
		}
	}
	return nil
}

func (g *Game) collectVisible(float64) error {
	rect, _ := g.view.Viewport()
	g.visible = g.world.Query(rect, g.visible[:0])
	return nil
}

func (g *Game) animate(dt float64) error {
	return aseprite.AdvanceAnimations(dt, g.libs, g.visible)
}

func (g *Game) followPlayer(dt float64) error {
	c, cOk := camera.Main(g.world)
	p, pOk := player.Player.First(g.world.ECS())
	if cOk && pOk {
		camera.Follow(dt, c, p)
	}
	if bounds, ok := mapBounds(g.world); cOk && ok {
		view, _ := g.view.Viewport()
		camera.Clamp(c, view, bounds)
	}
	return nil
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
