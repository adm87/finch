package display

import (
	"image/color"

	"github.com/adm87/finch/pkg/hexcolor"
)

type Config struct {
	Screen ScreenConfig `yaml:"screen"`
}

func DefaultConfig() Config {
	return Config{
		Screen: DefaultScreenConfig(),
	}
}

type ScreenConfig struct {
	Width           int            `yaml:"width"`
	Height          int            `yaml:"height"`
	ResolutionScale float64        `yaml:"resolution_scale"`
	ClearColor      hexcolor.Color `yaml:"clear_color"`
}

func DefaultScreenConfig() ScreenConfig {
	return ScreenConfig{
		Width:           400,
		Height:          300,
		ResolutionScale: 1,
		ClearColor:      hexcolor.Color(color.RGBA{A: 255}),
	}
}
