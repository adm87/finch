package engine

import (
	"fmt"
	"time"

	"github.com/hajimehoshi/ebiten/v2"

	"github.com/adm87/finch/pkg/application"
	"github.com/adm87/finch/pkg/errs"
)

// Run sets up the window from the app's config and runs the game loop.
func Run(app *application.Application, opts ...RunOption) error {
	if app == nil {
		return errs.ArgumentNil("app")
	}

	options := DefaultRunOptions()
	for _, opt := range opts {
		opt(options)
	}

	if err := start(app, options); err != nil {
		return err
	}

	cfg := app.Config()
	ebiten.SetWindowTitle(cfg.Title)
	ebiten.SetWindowSize(cfg.Screen.Width, cfg.Screen.Height)
	ebiten.SetFullscreen(cfg.Screen.Fullscreen)
	ebiten.SetWindowResizingMode(options.ResizeMode)
	ebiten.SetTPS(cfg.TickRate)

	return ebiten.RunGame(newShell(app, options, newClock(cfg.TickRate, time.Now)))
}

// start initializes the plugins, then runs the startup functions.
func start(app *application.Application, options *RunOptions) error {
	for _, startup := range options.Startup {
		if err := startup(app); err != nil {
			return fmt.Errorf("startup: %w", err)
		}
	}
	return nil
}
