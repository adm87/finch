package engine

import (
	"github.com/adm87/finch/pkg/logger"
)

type RunOptions struct {
	WorldDrawers   Drawers
	OverlayDrawers Drawers
	Logger         logger.Logger
	Update         []System
	FixedUpdate    []System
	LateUpdate     []System
}

type RunOption func(*RunOptions)

func DefaultRunOptions() *RunOptions {
	return &RunOptions{
		Logger: logger.N,
	}
}

// WithUpdate adds systems that run once per update, before the fixed steps.
func WithUpdate(systems ...System) RunOption {
	return func(ro *RunOptions) {
		ro.Update = append(ro.Update, systems...)
	}
}

// WithFixedUpdate adds systems that run once per fixed step at the tick rate.
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

// WithWorldDrawers adds drawers for the screen buffer, run when there's a camera.
func WithWorldDrawers(drawers ...Drawer) RunOption {
	return func(ro *RunOptions) {
		ro.WorldDrawers = append(ro.WorldDrawers, drawers...)
	}
}

func WithOverlayDrawers(drawers ...Drawer) RunOption {
	return func(ro *RunOptions) {
		ro.OverlayDrawers = append(ro.OverlayDrawers, drawers...)
	}
}

func WithLogger(logger logger.Logger) RunOption {
	return func(ro *RunOptions) {
		ro.Logger = logger
	}
}
