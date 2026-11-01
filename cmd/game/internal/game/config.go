package game

import (
	"github.com/adm87/finch/pkg/display"
	"github.com/adm87/finch/pkg/engine"
	"github.com/adm87/finch/pkg/resource"
)

type Config struct {
	Bundles map[string]Bundle `yaml:"bundles"`
	Startup []string          `yaml:"startup"` // bundles loaded before the first scene
	Engine  engine.Config     `yaml:"engine"`
	Screen  display.Config    `yaml:"screen"`
}

// Bundle is content loaded together from one source.
type Bundle struct {
	Source    resource.Source `yaml:"source"`
	Images    []string        `yaml:"images"`
	Maps      []string        `yaml:"maps"`
	Libraries []string        `yaml:"libraries"`
}

func DefaultConfig() *Config {
	return &Config{
		Screen: *display.DefaultConfig(),
		Engine: *engine.DefaultConfig(),
	}
}
