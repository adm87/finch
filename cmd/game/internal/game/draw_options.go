package game

import (
	"github.com/adm87/finch/pkg/hexcolor"
)

type DrawOptions struct {
	ClearColor hexcolor.Color
}

type DrawOption func(*DrawOptions)

func DefaultDrawOptions() *DrawOptions {
	return &DrawOptions{
		ClearColor: hexcolor.Color{A: 255},
	}
}

func WithClearColor(c hexcolor.Color) DrawOption {
	return func(opts *DrawOptions) {
		opts.ClearColor = c
	}
}
