package internal

import (
	"github.com/adm87/finch/cmd/finch-game/internal/game"
	"github.com/adm87/finch/pkg/engine"
)

func Boot() error {
	updater := &game.Updater{}
	return engine.Run(
		engine.WithDrawer(&game.Drawer{}),
		engine.WithUpdater(engine.UpdaterFunc(updater.Update)),
		engine.WithFixedUpdater(engine.UpdaterFunc(updater.FixedUpdate)),
		engine.WithLaterUpdater(engine.UpdaterFunc(updater.LaterUpdate)),
	)
}
