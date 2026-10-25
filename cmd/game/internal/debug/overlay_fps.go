package debug

import (
	"fmt"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"

	"github.com/adm87/finch/pkg/engine"
	"github.com/adm87/finch/pkg/world"
)

func OverlayFPS(w *world.World) engine.Drawer {
	return engine.DrawerFunc(func(target *ebiten.Image, _ *engine.DrawContext) error {
		if !GetDebug(w.ECS()).HasOverlay(OverlayTypeFPS) {
			return nil
		}
		fps := fmt.Sprintf("FPS: %.2f", ebiten.ActualFPS())
		ebitenutil.DebugPrintAt(target, fps, 10, 10)
		return nil
	})
}
