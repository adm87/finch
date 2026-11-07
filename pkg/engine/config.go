package engine

type Config struct {
	Ebiten EbitenConfig `yaml:"ebiten"`
	Clock  ClockConfig  `yaml:"clock"`
}

func DefaultConfig() Config {
	return Config{
		Ebiten: DefaultEbitenConfig(),
		Clock:  DefaultClockConfig(),
	}
}

type EbitenConfig struct {
	Window     WindowConfig `yaml:"window"`
	FullScreen bool         `yaml:"fullscreen"`
}

func DefaultEbitenConfig() EbitenConfig {
	return EbitenConfig{
		Window:     DefaultWindowConfig(),
		FullScreen: false,
	}
}

type WindowConfig struct {
	Title     string `yaml:"title"`
	Width     int    `yaml:"width"`
	Height    int    `yaml:"height"`
	Resizable bool   `yaml:"resizable"`
}

func DefaultWindowConfig() WindowConfig {
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

func DefaultClockConfig() ClockConfig {
	return ClockConfig{
		TickRate:  60,
		FixedRate: 60,
	}
}
