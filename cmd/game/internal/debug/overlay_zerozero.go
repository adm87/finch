package debug

import (
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"

	"github.com/adm87/finch/pkg/application"
)

func OverlayZeroZero(target *ebiten.Image, app *application.Application) error {
	x, y := app.View().WorldToScreen(0, 0)
	ebitenutil.DebugPrintAt(target, "0, 0", int(x), int(y))
	return nil
}
