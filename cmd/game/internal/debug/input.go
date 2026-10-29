package debug

import (
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"

	"github.com/adm87/finch/pkg/application"
	"github.com/adm87/finch/pkg/engine"
)

func PollInput() engine.System {
	return func(app *application.Application, _ float64) error {
		w := app.World()

		d := GetDebug(w.ECS())
		if inpututil.IsKeyJustPressed(ebiten.Key1) {
			d.SetOverlay(OverlayTypeEntities, !d.HasOverlay(OverlayTypeEntities))
		}
		if inpututil.IsKeyJustPressed(ebiten.Key2) {
			d.SetOverlay(OverlayTypeFPS, !d.HasOverlay(OverlayTypeFPS))
		}
		if inpututil.IsKeyJustPressed(ebiten.Key3) {
			d.SetOverlay(OverlayTypePartitions, !d.HasOverlay(OverlayTypePartitions))
		}
		if inpututil.IsKeyJustPressed(ebiten.Key4) {
			d.SetOverlay(OverlayTypeZeroZero, !d.HasOverlay(OverlayTypeZeroZero))
		}

		return nil
	}
}
