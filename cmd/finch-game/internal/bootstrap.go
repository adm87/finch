package internal

import (
	"github.com/adm87/finch/cmd/finch-game/internal/game/draw"
	"github.com/adm87/finch/cmd/finch-game/internal/game/update"
	"github.com/adm87/finch/pkg/display"
	"github.com/adm87/finch/pkg/engine"
)

func Boot() error {
	screen := display.NewScreen(400, 300)

	drawer := draw.New(screen)
	updater := update.New()

	return engine.Run(
		engine.WithDrawer(drawer.Pipeline()),
		engine.WithUpdater(updater.UpdatePhase()),
		engine.WithDisplay(screen),
	)
}
