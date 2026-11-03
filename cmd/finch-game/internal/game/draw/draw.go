package draw

import (
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"

	"github.com/adm87/finch/pkg/display"
	"github.com/adm87/finch/pkg/engine"
)

type Drawer struct {
	screen *display.Screen
}

func New(screen *display.Screen) *Drawer {
	return &Drawer{
		screen: screen,
	}
}

func (d *Drawer) Pipeline() engine.Drawer {
	return engine.DrawPipeline(
		d.drawWorld(),
		// TODO: Add post-processing pass
		d.screen.Present(),
		// TODO: Add UI pass
		// TODO: Add debug pass
	)
}

func (d *Drawer) drawWorld() engine.DrawerFunc {
	return func(target *ebiten.Image) {
		visible := d.screen.Visible()

		minX, minY := visible.Min()
		maxX, maxY := visible.Max()

		vector.FillRect(target, float32(minX), float32(minY), 20, 20, color.RGBA{R: 255, A: 255}, false)
		vector.FillRect(target, float32(maxX-20), float32(minY), 20, 20, color.RGBA{G: 255, A: 255}, false)
		vector.FillRect(target, float32(minX), float32(maxY-20), 20, 20, color.RGBA{B: 255, A: 255}, false)
		vector.FillRect(target, float32(maxX-20), float32(maxY-20), 20, 20, color.RGBA{R: 255, G: 255, A: 255}, false)
	}
}
