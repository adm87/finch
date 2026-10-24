package debug

import (
	"fmt"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"

	"github.com/adm87/finch/pkg/application"
)

func OverlayFPS(target *ebiten.Image, _ *application.Application) error {
	fps := fmt.Sprintf("FPS: %.2f", ebiten.ActualFPS())
	ebitenutil.DebugPrintAt(target, fps, 10, 10)
	return nil
}
