package game

import (
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
	x, y := t.Position()

	if ebiten.IsKeyPressed(ebiten.KeyA) {
		x -= 5
	}
	if ebiten.IsKeyPressed(ebiten.KeyD) {
		x += 5
	}
	if ebiten.IsKeyPressed(ebiten.KeyW) {
		y -= 5
	}
	if ebiten.IsKeyPressed(ebiten.KeyS) {
		y += 5
	}

	t.SetPosition(x, y)
	return nil
}
