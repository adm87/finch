package engine

type WindowConfig struct {
	Title     string `yaml:"title"`
	Width     int    `yaml:"width"`
	Height    int    `yaml:"height"`
	Resizable bool   `yaml:"resizable"`
}

type ClockConfig struct {
	TickRate  int `yaml:"tick_rate"`  // updates per second
	FixedRate int `yaml:"fixed_rate"` // fixed steps per second
}

// Config holds the engine settings Run applies at startup.
type Config struct {
	Window     WindowConfig `yaml:"window"`
	Clock      ClockConfig  `yaml:"clock"`
	Fullscreen bool         `yaml:"fullscreen"`
}

// DefaultConfig is the config Run uses when none is given. Decode config files
// onto it so missing keys keep their defaults.
func DefaultConfig() Config {
	return Config{
		Window: WindowConfig{
			Title:     "Untitled",
			Width:     800,
			Height:    600,
			Resizable: true,
		},
		Clock: ClockConfig{
			TickRate:  60,
			FixedRate: 60,
		},
	}
}
