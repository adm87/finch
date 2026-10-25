package tiled

import (
	"github.com/adm87/finch/pkg/errs"
)

type MapConfig struct {
	Name       string `yaml:"name"`
	FileSystem string `yaml:"filesystem"`
	TMX        string `yaml:"tmx"`
}

type Config struct {
	Maps map[string]MapConfig `yaml:"maps"`
}

func (c Config) Validate() error {
	for _, m := range c.Maps {
		if m.Name == "" {
			return errs.Argument("name", "map name is required")
		}
		if m.FileSystem == "" {
			return errs.Argument("filesystem", "map filesystem is required")
		}
		if m.TMX == "" {
			return errs.Argument("tmx", "map tmx is required")
		}
	}
	return nil
}
