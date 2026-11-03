package engine

import (
	"github.com/hajimehoshi/ebiten/v2"
)

func Run(opts ...Option) error {
	o := defaultOptions()
	for _, opt := range opts {
		opt(&o)
	}

	initializeEbiten(o.config)

	return ebiten.RunGame(&shell{
		clock: newClock(
			o.config.Clock.TickRate,
			o.config.Clock.FixedRate,
		),
		display:      o.display,
		drawer:       o.drawer,
		updater:      o.updater,
		fixedUpdater: o.fixedUpdater,
		laterUpdater: o.laterUpdater,
	})
}

func initializeEbiten(config Config) {
	ebiten.SetTPS(max(config.Clock.TickRate, 1))
	ebiten.SetWindowTitle(config.Window.Title)
	ebiten.SetWindowSize(config.Window.Width, config.Window.Height)

	mode := ebiten.WindowResizingModeDisabled
	if config.Window.Resizable {
		mode = ebiten.WindowResizingModeEnabled
	}
	ebiten.SetWindowResizingMode(mode)
	ebiten.SetFullscreen(config.Fullscreen)
}
