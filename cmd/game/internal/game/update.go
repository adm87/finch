package game

import (
	"math"

	"github.com/adm87/finch/pkg/ecs/components/transform"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/yohamta/donburi"
)

type GameUpdate struct {
	camera *donburi.Entry
}

func NewGameUpdate(camera *donburi.Entry) *GameUpdate {
	return &GameUpdate{
		camera: camera,
	}
}

func (gu *GameUpdate) Update() error {
	if inpututil.IsKeyJustPressed(ebiten.KeyEscape) {
		return ebiten.Termination
	}

	t := transform.TransformComponent.Get(gu.camera)

	sx, sy := t.Scale()
	if ebiten.IsKeyPressed(ebiten.KeyUp) {
		sx *= 0.95
		sy *= 0.95
	}
	if ebiten.IsKeyPressed(ebiten.KeyDown) {
		sx *= 1.05
		sy *= 1.05
	}
	t.SetScale(sx, sy)

	rot := t.Rotation()
	if ebiten.IsKeyPressed(ebiten.KeyQ) {
		rot += 0.05
	}
	if ebiten.IsKeyPressed(ebiten.KeyE) {
		rot -= 0.05
	}
	t.SetRotation(rot)

	x, y := t.Position()

	var dx, dy float64
	if ebiten.IsKeyPressed(ebiten.KeyA) {
		dx -= 5
	}
	if ebiten.IsKeyPressed(ebiten.KeyD) {
		dx += 5
	}
	if ebiten.IsKeyPressed(ebiten.KeyW) {
		dy -= 5
	}
	if ebiten.IsKeyPressed(ebiten.KeyS) {
		dy += 5
	}

	if dx != 0 || dy != 0 {
		cos, sin := math.Cos(rot), math.Sin(rot)
		rdx := dx*cos - dy*sin
		rdy := dx*sin + dy*cos

		x += rdx * sx
		y += rdy * sy
	}

	t.SetPosition(x, y)
	return nil
}
