package game

type ScreenConfig struct {
	Width           int     `yaml:"width"`
	Height          int     `yaml:"height"`
	ResolutionScale float64 `yaml:"resolution_scale"`
}

type Config struct {
	Screen ScreenConfig `yaml:"screen"`
}

// DefaultConfig is the game's config before game.yaml is applied. Decode the
// file onto it so missing keys keep their defaults.
func DefaultConfig() Config {
	return Config{
		Screen: ScreenConfig{
			Width:           400,
			Height:          300,
			ResolutionScale: 1,
		},
	}
}
