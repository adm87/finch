package engine

import (
	"time"

	"github.com/hajimehoshi/ebiten/v2"

	"github.com/adm87/finch/pkg/display"
	"github.com/adm87/finch/pkg/world"
)

// Run sets up the window and runs the game loop. A nil config uses DefaultConfig.
func Run(config *Config, w *world.World, screen *display.Screen, view *display.View, opts ...RunOption) error {
	cfg := *DefaultConfig()
	if config != nil {
		cfg = *config
	}
	if err := cfg.Validate(); err != nil {
		return err
	}

	options := DefaultRunOptions()
	for _, opt := range opts {
		opt(options)
	}

	resizing := ebiten.WindowResizingModeDisabled
	if cfg.Window.Resizable {
		resizing = ebiten.WindowResizingModeEnabled
	}

	ebiten.SetWindowTitle(cfg.Title)
	ebiten.SetWindowSize(cfg.Window.Width, cfg.Window.Height)
	ebiten.SetFullscreen(cfg.Window.Fullscreen)
	ebiten.SetWindowResizingMode(resizing)
	ebiten.SetTPS(cfg.TickRate)

	return ebiten.RunGame(newShell(w, screen, view, cfg, options, newClock(cfg.TickRate, time.Now)))
}
