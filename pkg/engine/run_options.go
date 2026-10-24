package engine

import (
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"

	"github.com/adm87/finch/pkg/logger"
)

type RunOptions struct {
	Update      []System
	FixedUpdate []System
	LateUpdate  []System
	Drawer      Drawer
	// Overlay draws over the window after the screen buffer, in window coordinates.
	Overlay    Drawer
	ClearColor color.Color
	Logger     logger.Logger
	ResizeMode ebiten.WindowResizingModeType
}

type RunOption func(*RunOptions)

func DefaultRunOptions() *RunOptions {
	return &RunOptions{
		Drawer:     DrawerFunc(noopDraw),
		Overlay:    DrawerFunc(noopDraw),
		ClearColor: color.Black,
		Logger:     logger.N,
		ResizeMode: ebiten.WindowResizingModeDisabled,
	}
}

// WithUpdate adds systems that run once per update, before the fixed steps.
func WithUpdate(systems ...System) RunOption {
	return func(ro *RunOptions) {
		ro.Update = append(ro.Update, systems...)
	}
}

// WithFixedUpdate adds systems that run once per fixed step at the app's tick rate.
func WithFixedUpdate(systems ...System) RunOption {
	return func(ro *RunOptions) {
		ro.FixedUpdate = append(ro.FixedUpdate, systems...)
	}
}

// WithLateUpdate adds systems that run once per update, after the fixed steps.
func WithLateUpdate(systems ...System) RunOption {
	return func(ro *RunOptions) {
		ro.LateUpdate = append(ro.LateUpdate, systems...)
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
