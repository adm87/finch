package debug

import (
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"

	"github.com/adm87/finch/pkg/engine"
	"github.com/adm87/finch/pkg/world"
)

func OverlayZeroZero(w *world.World) engine.Drawer {
	return engine.DrawerFunc(func(target *ebiten.Image, ctx *engine.DrawContext) error {
		if !GetDebug(w.ECS()).HasOverlay(OverlayTypeZeroZero) {
			return nil
		}
		x, y := ctx.Matrix.Apply(0, 0)
		ebitenutil.DebugPrintAt(target, "0, 0", int(x), int(y))
		return nil
	})
}
