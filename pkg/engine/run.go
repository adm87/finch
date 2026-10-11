package engine

import "github.com/hajimehoshi/ebiten/v2"

func Run(opts ...RunOption) error {
	options := DefaultRunOptions()
	for _, opt := range opts {
		opt(options)
	}

	sh := shell{
		update: options.Update,
		draw:   options.Draw,
		layout: options.Layout,
		logger: options.Logger,
	}

	ebiten.SetWindowSize(options.WindowWidth, options.WindowHeight)
	ebiten.SetFullscreen(options.Fullscreen)
	ebiten.SetWindowResizingMode(options.ResizeMode)

	return ebiten.RunGame(&sh)
}
