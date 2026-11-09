package engine

// Config holds the engine settings Run applies at startup.
type Config struct {
	Ebiten EbitenConfig `yaml:"ebiten"`
	Clock  ClockConfig  `yaml:"clock"`
}

// DefaultConfig is the config Run uses when none is given. Decode config files
// onto it so missing keys keep their defaults.
func DefaultConfig() Config {
	return Config{
		Ebiten: defaultEbitenConfig(),
		Clock:  defaultClockConfig(),
	}
}

type EbitenConfig struct {
	Window     WindowConfig `yaml:"window"`
	Fullscreen bool         `yaml:"fullscreen"`
}

func defaultEbitenConfig() EbitenConfig {
	return EbitenConfig{
		Window:     defaultWindowConfig(),
		Fullscreen: false,
	}
}

type WindowConfig struct {
	Title     string `yaml:"title"`
	Width     int    `yaml:"width"`
	Height    int    `yaml:"height"`
	Resizable bool   `yaml:"resizable"`
}

func defaultWindowConfig() WindowConfig {
	return WindowConfig{
		Title:     "Untitled",
		Width:     800,
		Height:    600,
		Resizable: true,
	}
}

type ClockConfig struct {
	TickRate  int `yaml:"tick_rate"`  // updates per second
	FixedRate int `yaml:"fixed_rate"` // fixed steps per second
}

func defaultClockConfig() ClockConfig {
	return ClockConfig{
		TickRate:  60,
		FixedRate: 60,
	}
}
