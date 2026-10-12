package game

import (
	"github.com/adm87/finch/pkg/types"
)

type DrawOptions struct {
	ClearColor types.HexColor
}

type DrawOption func(*DrawOptions)

func DefaultDrawOptions() *DrawOptions {
	return &DrawOptions{
		ClearColor: types.HexColor{A: 255},
	}
}

func WithClearColor(c types.HexColor) DrawOption {
	return func(opts *DrawOptions) {
		opts.ClearColor = c
	}
}
