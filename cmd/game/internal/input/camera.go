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
}
