package display

import (
	"image/color"

	"github.com/adm87/finch/pkg/hexcolor"
)

// Config holds the display settings.
type Config struct {
	Screen ScreenConfig `yaml:"screen"`
}

// DefaultConfig is the display config before any file is applied. Decode
// config files onto it so missing keys keep their defaults.
func DefaultConfig() Config {
	return Config{
		Screen: defaultScreenConfig(),
	}
}

type ScreenConfig struct {
	Width           int            `yaml:"width"`
	Height          int            `yaml:"height"`
	ResolutionScale float64        `yaml:"resolution_scale"`
	ClearColor      hexcolor.Color `yaml:"clear_color"`
}

func defaultScreenConfig() ScreenConfig {
	return ScreenConfig{
		Width:           400,
		Height:          300,
		ResolutionScale: 1,
		ClearColor:      hexcolor.Color(color.RGBA{A: 255}),
	}
}
