package game

import "image/color"

type DrawOptions struct {
	ClearColor color.RGBA
}

type DrawOption func(*DrawOptions)

func DefaultDrawOptions() *DrawOptions {
	return &DrawOptions{
		ClearColor: color.RGBA{A: 255},
	}
}

func WithClearColor(c color.RGBA) DrawOption {
	return func(opts *DrawOptions) {
		opts.ClearColor = c
	}
}
