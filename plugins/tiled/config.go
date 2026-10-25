package tiled

import (
	"errors"
	"fmt"

	"github.com/adm87/finch/pkg/assets"
	"github.com/adm87/finch/pkg/errs"
)

type MapConfig struct {
	FileSystem assets.FileSystem `yaml:"filesystem"`
	TMX        assets.Asset      `yaml:"tmx"`
}

// Config lists maps by name.
type Config struct {
	Maps map[string]MapConfig `yaml:"maps"`
}

func (c Config) Validate() error {
	var errList []error
	for name, m := range c.Maps {
		if m.FileSystem == "" {
			errList = append(errList, errs.Argument("filesystem", fmt.Sprintf("map %q needs a file system", name)))
		}
		if m.TMX == "" {
			errList = append(errList, errs.Argument("tmx", fmt.Sprintf("map %q needs a .tmx file", name)))
		}
	}
	return errors.Join(errList...)
}
