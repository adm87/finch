package devinfo

import (
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/vector"
	"github.com/yohamta/donburi"

	"github.com/adm87/finch/cmd/finch-game/internal/game/camera"
	"github.com/adm87/finch/pkg/display"
	"github.com/adm87/finch/pkg/engine"
)

// DrawWorldZero marks world (0, 0) with a small cross while it's on screen.
func DrawWorldZero(world donburi.World, screen *display.Screen) engine.DrawPass {
	const size = 5
	return func(inner engine.Drawer) engine.Drawer {
		return engine.DrawFunc(func(target *ebiten.Image) {
			inner.Draw(target)

			view := camera.View(world, screen.Visible())
			if !view.Bounds.Contains(0, 0) {
				return
			}

			x, y := screen.BufferToWindow(view.Matrix.Apply(0, 0))
			fx, fy := float32(x), float32(y)

			vector.StrokeLine(target, fx-size, fy, fx+size, fy, 1, color.White, false)
			vector.StrokeLine(target, fx, fy-size, fx, fy+size, 1, color.White, false)

			ebitenutil.DebugPrintAt(target, "(0, 0)", int(fx+size), int(fy+size))
		})
	}
}
