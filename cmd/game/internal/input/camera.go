package input

import (
	"math"

	"github.com/adm87/finch/pkg/ecs/components"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/yohamta/donburi"
)

const (
	camMoveSpeed = 300.0
	camRotSpeed  = 3.0
	camZoomRate  = 1.05
)

func UpdateCamera(dt float64, c *donburi.Entry) {
	t := components.TransformComponent.Get(c)

	// Zoom
	sx, sy := t.Scale()
	zoom := 0.0
	if ebiten.IsKeyPressed(ebiten.KeyUp) {
		zoom -= 1
	}
	if ebiten.IsKeyPressed(ebiten.KeyDown) {
		zoom += 1
	}
	if zoom != 0 {
		f := math.Pow(camZoomRate, zoom*dt*60)
		sx *= f
		sy *= f
	}
	t.SetScale(sx, sy)

	// Rotation
	rot := t.Rotation()
	if ebiten.IsKeyPressed(ebiten.KeyQ) {
		rot += camRotSpeed * dt
	}
	if ebiten.IsKeyPressed(ebiten.KeyE) {
		rot -= camRotSpeed * dt
	}
	t.SetRotation(rot)

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

		cos, sin := math.Cos(rot), math.Sin(rot)
		rdx := dx*cos - dy*sin
		rdy := dx*sin + dy*cos

		x, y := t.Position()
		x += rdx * sx * camMoveSpeed * dt
		y += rdy * sy * camMoveSpeed * dt
		t.SetPosition(x, y)
	}
}
