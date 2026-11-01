package game

import (
	"github.com/adm87/finch/pkg/display"
	"github.com/adm87/finch/pkg/engine"
)

type Config struct {
	Preload []string       `yaml:"preload"` // embedded images loaded at startup
	Engine  engine.Config        `yaml:"engine"`
	Screen  display.Config       `yaml:"screen"`
}

func DefaultConfig() *Config {
	return &Config{
		Screen: *display.DefaultConfig(),
		Engine: *engine.DefaultConfig(),
	}
}
