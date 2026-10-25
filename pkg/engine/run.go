package engine

import (
	"fmt"
	"time"

	"github.com/hajimehoshi/ebiten/v2"

	"github.com/adm87/finch/pkg/application"
	"github.com/adm87/finch/pkg/errs"
)

// Run sets up the window and runs the game loop. A nil config uses DefaultConfig.
func Run(app *application.Application, config *Config, opts ...RunOption) error {
	if app == nil {
		return errs.ArgumentNil("app")
	}

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

	if err := start(app, options); err != nil {
		return err
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

	return ebiten.RunGame(newShell(app, cfg, options, newClock(cfg.TickRate, time.Now)))
}

// start runs the startup functions in order.
func start(app *application.Application, options *RunOptions) error {
	for _, startup := range options.Startup {
		if err := startup(app); err != nil {
			return fmt.Errorf("startup: %w", err)
		}
	}
	return nil
}
