package game

import (
	"github.com/adm87/finch/pkg/assets"
	"github.com/adm87/finch/pkg/display"
	"github.com/adm87/finch/pkg/engine"
)

type Config struct {
	Bundles []assets.AssetBundle `yaml:"bundles"`
	Engine  engine.Config        `yaml:"engine"`
	Screen  display.Config       `yaml:"screen"`
}

func DefaultConfig() *Config {
	return &Config{
		Screen: *display.DefaultConfig(),
		Engine: *engine.DefaultConfig(),
	}
}
