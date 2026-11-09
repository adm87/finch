package display

import "github.com/adm87/finch/pkg/hexcolor"

// ScreenConfig sets the screen buffer's size, its resolution scale and the
// colour it's cleared to each frame.
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
		ClearColor:      hexcolor.Color{A: 255},
	}
}

// Config holds the display settings.
type Config struct {
	Screen ScreenConfig `yaml:"screen"`
}

// DefaultConfig is the display config before a file is loaded. Keys the
// file leaves out keep these values.
func DefaultConfig() Config {
	return Config{
		Screen: defaultScreenConfig(),
	}
}
