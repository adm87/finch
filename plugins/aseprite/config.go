package aseprite

import (
	"errors"
	"fmt"

	"github.com/adm87/finch/pkg/errs"
)

type Config struct {
	Libraries map[string]LibraryConfig `yaml:"libraries"`
}

func DefaultConfig() *Config {
	return &Config{Libraries: make(map[string]LibraryConfig)}
}

func (c Config) Validate() error {
	var errList []error
	for name, lib := range c.Libraries {
		if err := lib.Validate(); err != nil {
			errList = append(errList, errs.Argument("config", fmt.Sprintf("library %q", name)).Wrap(err))
		}
	}
	return errors.Join(errList...)
}

type LibraryConfig struct {
	Atlas string `yaml:"atlas"`
	Image string `yaml:"image"`
}

func (lc LibraryConfig) Validate() error {
	if lc.Atlas == "" {
		return errs.Argument("atlas", "atlas cannot be empty")
	}
	if lc.Image == "" {
		return errs.Argument("image", "image cannot be empty")
	}
	return nil
}
