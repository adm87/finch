package game

import (
	"github.com/adm87/finch/pkg/display"
	"github.com/adm87/finch/pkg/engine"
)

type Config struct {
	Bundle  []string       `yaml:"bundle"`
	Engine  engine.Config  `yaml:"engine"`
	Display display.Config `yaml:"display"`
}

func DefaultConfig() Config {
	return Config{
		Engine:  engine.DefaultConfig(),
		Display: display.DefaultConfig(),
		Bundle:  []string{},
	}
}
