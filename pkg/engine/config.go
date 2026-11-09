package engine

// WindowConfig sets the window's title and starting size.
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

// EbitenConfig holds the settings Run passes to Ebitengine.
type EbitenConfig struct {
	Window     WindowConfig `yaml:"window"`
	Fullscreen bool         `yaml:"fullscreen"`
}

func defaultEbitenConfig() EbitenConfig {
	return EbitenConfig{
		Window: defaultWindowConfig(),
	}
}

// ClockConfig sets how often the update and fixed phases run.
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

// Config holds the engine settings Run applies at startup.
type Config struct {
	Ebiten EbitenConfig `yaml:"ebiten"`
	Clock  ClockConfig  `yaml:"clock"`
}

// DefaultConfig is what Run uses when it's given no config. Keys a config
// file leaves out keep these values.
func DefaultConfig() Config {
	return Config{
		Ebiten: defaultEbitenConfig(),
		Clock:  defaultClockConfig(),
	}
}
