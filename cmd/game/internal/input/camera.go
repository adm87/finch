package input

import (
	"math"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/yohamta/donburi"

	"github.com/adm87/finch/pkg/transform"
)

const (
	camMoveSpeed = 300.0
	camRotSpeed  = 3.0
	camZoomRate  = 1.05
)

func UpdateCamera(dt float64, c *donburi.Entry) {
	t := transform.TransformComponent.Get(c)

	// Zoom
	sx, sy := t.Scale()
	zoom := 0.0
	if ebiten.IsKeyPressed(ebiten.KeyUp) {
		zoom--
	}
	if ebiten.IsKeyPressed(ebiten.KeyDown) {
		zoom++
	}
	if zoom != 0 {
		f := math.Pow(camZoomRate, zoom*dt*60)
		sx *= f
		sy *= f
	}

	t.SetScale(sx, sy)
}
