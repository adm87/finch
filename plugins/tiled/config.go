package tiled

import (
	"errors"
	"fmt"

	"github.com/adm87/finch/pkg/assets"
	"github.com/adm87/finch/pkg/errs"
)

type MapConfig struct {
	TMX assets.Asset `yaml:"tmx"`
}

// Config lists maps by name.
type Config struct {
	Maps map[string]MapConfig `yaml:"maps"`
}

func DefaultConfig() *Config {
	return &Config{Maps: make(map[string]MapConfig)}
}

func (c Config) Validate() error {
	var errList []error
	for name, m := range c.Maps {
		if m.TMX == "" {
			errList = append(errList, errs.Argument("tmx", fmt.Sprintf("map %q needs a .tmx file", name)))
		}
	}
	return errors.Join(errList...)
}
