package game

import (
	"github.com/adm87/finch/pkg/display"
	"github.com/adm87/finch/pkg/engine"
)

type Config struct {
	Engine  engine.Config  `yaml:"engine"`
	Display display.Config `yaml:"display"`
}

// DefaultConfig is the config before game.yaml is loaded. Keys the file
// leaves out keep these values.
func DefaultConfig() Config {
	return Config{
		Engine:  engine.DefaultConfig(),
		Display: display.DefaultConfig(),
	}
}
