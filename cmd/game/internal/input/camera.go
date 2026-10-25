package input

import (
	"math"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/yohamta/donburi"

	"github.com/adm87/finch/cmd/game/internal/camera"
)

const camZoomRate = 1.05

func UpdateCamera(dt float64, c *donburi.Entry) {
	var dir float64
	if ebiten.IsKeyPressed(ebiten.KeyUp) {
		dir++
	}
	if ebiten.IsKeyPressed(ebiten.KeyDown) {
		dir--
	}
	if dir != 0 {
		camera.ZoomBy(c, math.Pow(camZoomRate, dir*dt*60))
	}
}
