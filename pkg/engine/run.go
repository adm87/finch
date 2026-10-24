package engine

import (
	"github.com/hajimehoshi/ebiten/v2"

	"github.com/adm87/finch/pkg/application"
	"github.com/adm87/finch/pkg/errs"
)

// Run sizes the window from the app's config and runs the game loop.
func Run(app *application.Application, opts ...RunOption) error {
	if app == nil {
		return errs.ArgumentNil("app")
	}

	options := DefaultRunOptions()
	for _, opt := range opts {
		opt(options)
	}

	cfg := app.Config()
	ebiten.SetWindowSize(cfg.Screen.Width, cfg.Screen.Height)
	ebiten.SetFullscreen(cfg.Screen.Fullscreen)
	ebiten.SetWindowResizingMode(options.ResizeMode)

	return ebiten.RunGame(&shell{
		app:        app,
		updater:    options.Updater,
		drawer:     options.Drawer,
		overlay:    options.Overlay,
		clearColor: options.ClearColor,
		logger:     options.Logger,
	})
}
