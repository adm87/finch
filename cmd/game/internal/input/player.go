package input

import (
	"math"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/yohamta/donburi"

	"github.com/adm87/finch/pkg/ecs/components"
	"github.com/adm87/finch/pkg/ecs/world"
)

const (
	playerMoveSpeed = 100.0
)

func UpdatePlayer(dt float64, entry *donburi.Entry, w world.World) error {
	t := components.TransformComponent.Get(entry)

	// Movement
	var dx, dy float64
	if ebiten.IsKeyPressed(ebiten.KeyA) {
		dx -= 1
	}
	if ebiten.IsKeyPressed(ebiten.KeyD) {
		dx += 1
	}
	if ebiten.IsKeyPressed(ebiten.KeyW) {
		dy -= 1
	}
	if ebiten.IsKeyPressed(ebiten.KeyS) {
		dy += 1
	}

	if dx != 0 || dy != 0 {
		l := math.Hypot(dx, dy)
		dx, dy = dx/l, dy/l

		x, y := t.Position()
		x += dx * playerMoveSpeed * dt
		y += dy * playerMoveSpeed * dt
		t.SetPosition(x, y)

		w.Update(entry)
	}

	return nil
}
