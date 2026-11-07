// Package engine runs the game loop: update phases, drawing and layout.
package engine

import (
	"github.com/hajimehoshi/ebiten/v2"
)

// Run opens the window and runs the game until an updater returns an error.
// Return ebiten.Termination to quit cleanly.
func Run(opts ...Option) error {
	o := defaultOptions()
	for _, opt := range opts {
		opt(&o)
	}

	initializeEbiten(o.config.Ebiten, o.config.Clock.TickRate)

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

func initializeEbiten(config EbitenConfig, tickRate int) {
	ebiten.SetTPS(max(tickRate, 1))
	ebiten.SetWindowTitle(config.Window.Title)
	ebiten.SetWindowSize(config.Window.Width, config.Window.Height)

	mode := ebiten.WindowResizingModeDisabled
	if config.Window.Resizable {
		mode = ebiten.WindowResizingModeEnabled
	}
	ebiten.SetWindowResizingMode(mode)
	ebiten.SetFullscreen(config.FullScreen)
}
