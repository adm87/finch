package debug

import (
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"

	"github.com/adm87/finch/pkg/engine"
)

func OverlayZeroZero(target *ebiten.Image, ctx *engine.DrawContext) error {
	if ctx.View.Area() == 0 {
		return nil
	}
	x, y := ctx.Matrix.Apply(0, 0)
	ebitenutil.DebugPrintAt(target, "0, 0", int(x), int(y))
	return nil
}
