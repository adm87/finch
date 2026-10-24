package debug

import (
	"fmt"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"

	"github.com/adm87/finch/pkg/application"
)

// DrawDebugOverlay labels the world origin and shows the FPS.
func DrawDebugOverlay(target *ebiten.Image, app *application.Application) error {
	x, y := app.View().WorldToScreen(0, 0)
	ebitenutil.DebugPrintAt(target, "0, 0", int(x), int(y))

	fps := fmt.Sprintf("FPS: %.2f", ebiten.ActualFPS())
	ebitenutil.DebugPrintAt(target, fps, 10, 10)
	return nil
}
