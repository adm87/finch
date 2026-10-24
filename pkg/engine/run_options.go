package engine

import (
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"

	"github.com/adm87/finch/pkg/logger"
)

type RunOptions struct {
	Updater Updater
	Drawer  Drawer
	// Overlay draws over the window after the screen buffer, in window coordinates.
	Overlay    Drawer
	ClearColor color.Color
	Logger     logger.Logger
	ResizeMode ebiten.WindowResizingModeType
}

type RunOption func(*RunOptions)

func DefaultRunOptions() *RunOptions {
	return &RunOptions{
		Updater:    UpdaterFunc(noopUpdate),
		Drawer:     DrawerFunc(noopDraw),
		Overlay:    DrawerFunc(noopDraw),
		ClearColor: color.Black,
		Logger:     logger.N,
		ResizeMode: ebiten.WindowResizingModeDisabled,
	}
}

func WithUpdater(updater Updater) RunOption {
	return func(ro *RunOptions) {
		ro.Updater = updater
	}
}

func WithDrawer(drawer Drawer) RunOption {
	return func(ro *RunOptions) {
		ro.Drawer = drawer
	}
}

func WithOverlay(overlay Drawer) RunOption {
	return func(ro *RunOptions) {
		ro.Overlay = overlay
	}
}

func WithClearColor(c color.Color) RunOption {
	return func(ro *RunOptions) {
		ro.ClearColor = c
	}
}

func WithLogger(logger logger.Logger) RunOption {
	return func(ro *RunOptions) {
		ro.Logger = logger
	}
}

func WithResizeMode(resizeMode ebiten.WindowResizingModeType) RunOption {
	return func(ro *RunOptions) {
		ro.ResizeMode = resizeMode
	}
}
