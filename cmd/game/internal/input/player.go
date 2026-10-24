package input

import (
	"math"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/yohamta/donburi"

	"github.com/adm87/finch/pkg/transform"
	"github.com/adm87/finch/pkg/world"
	"github.com/adm87/finch/plugins/aseprite"
)

const (
	playerMoveSpeed = 100.0
)

func UpdatePlayer(dt float64, entry *donburi.Entry, w *world.World) error {
	t := transform.TransformComponent.Get(entry)

	// Movement
	var dx, dy float64
	if ebiten.IsKeyPressed(ebiten.KeyA) {
		dx--
	}
	if ebiten.IsKeyPressed(ebiten.KeyD) {
		dx++
	}
	if ebiten.IsKeyPressed(ebiten.KeyW) {
		dy--
	}
	if ebiten.IsKeyPressed(ebiten.KeyS) {
		dy++
	}

	// This goes in late update, put here for now
	if anim, ok := aseprite.GetAnimator(entry); ok {
		if dx != 0 || dy != 0 {
			sx, sy := t.Scale()
			if dx < 0 {
				t.SetScale(-math.Abs(sx), sy)
			} else if dx > 0 {
				t.SetScale(math.Abs(sx), sy)
			}
			anim.SetAnimation("Run")
		} else {
			anim.SetAnimation("Idle")
		}
	}

	if dx != 0 || dy != 0 {

		l := math.Hypot(dx, dy)
		dx, dy = dx/l, dy/l

		x, y := t.Position()
		x += dx * playerMoveSpeed * dt
		y += dy * playerMoveSpeed * dt
		t.SetPosition(x, y)

		return w.Update(entry)
	}

	return nil
}
