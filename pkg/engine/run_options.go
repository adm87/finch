package engine

import "github.com/adm87/finch/pkg/logger"

type RunOptions struct {
	Update GameUpdate
	Draw   GameDraw
	Layout GameLayout

	Logger logger.Logger

	WindowWidth  int
	WindowHeight int
	Fullscreen   bool
}

type RunOption func(*RunOptions)

func DefaultRunOptions() *RunOptions {
	return &RunOptions{
		Update:       &noopUpdate{},
		Draw:         &noopDraw{},
		Layout:       &noopLayout{},
		Logger:       logger.N,
		WindowWidth:  800,
		WindowHeight: 600,
		Fullscreen:   false,
	}
}

func WithUpdate(update GameUpdate) RunOption {
	return func(ro *RunOptions) {
		ro.Update = update
	}
}

func WithDraw(draw GameDraw) RunOption {
	return func(ro *RunOptions) {
		ro.Draw = draw
	}
}

func WithLayout(layout GameLayout) RunOption {
	return func(ro *RunOptions) {
		ro.Layout = layout
	}
}

func WithLogger(logger logger.Logger) RunOption {
	return func(ro *RunOptions) {
		ro.Logger = logger
	}
}

func WithWindowSize(width, height int) RunOption {
	return func(ro *RunOptions) {
		ro.WindowWidth = width
		ro.WindowHeight = height
	}
}

func WithFullscreen(fullscreen bool) RunOption {
	return func(ro *RunOptions) {
		ro.Fullscreen = fullscreen
	}
}
