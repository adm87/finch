package engine

type Config struct {
	Window     WindowConfig `yaml:"window"`
	Clock      ClockConfig  `yaml:"clock"`
	Fullscreen bool         `yaml:"fullscreen"`
}

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
