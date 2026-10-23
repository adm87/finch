package aseprite

import (
	"errors"
	"fmt"

	"github.com/adm87/finch/pkg/assets"
	"github.com/adm87/finch/pkg/errs"
)

type Config struct {
	Libraries map[string]LibraryConfig `yaml:"libraries"`
}

// Validate checks every library config and reports all problems found.
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
	Atlas      assets.Asset `yaml:"atlas"`
	Image      assets.Asset `yaml:"image"`
	FileSystem string       `yaml:"filesystem"`
}

func (lc LibraryConfig) Validate() error {
	if lc.Atlas.String() == "" {
		return errs.Argument("atlas", "atlas cannot be empty")
	}
	if lc.Image.String() == "" {
		return errs.Argument("image", "image cannot be empty")
	}
	if lc.FileSystem == "" {
		return errs.Argument("filesystem", "filesystem cannot be empty")
	}
	return nil
}

func (lc LibraryConfig) Assets() []assets.Asset {
	return []assets.Asset{lc.Atlas, lc.Image}
}
