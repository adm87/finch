package aseprite

import (
	"fmt"

	"github.com/adm87/finch/pkg/resources"
	"github.com/adm87/finch/pkg/types/errs"
)

type Config struct {
	Libraries map[string]LibraryConfig `yaml:"libraries"`
}

func (c Config) Validate() error {
	for name, lib := range c.Libraries {
		if err := lib.Validate(); err != nil {
			return errs.Argument("libraries", fmt.Sprintf("library %q: %v", name, err))
		}
	}
	return nil
}

type LibraryConfig struct {
	Atlas      resources.Resource `yaml:"atlas"`
	Image      resources.Resource `yaml:"image"`
	FileSystem string             `yaml:"filesystem"`
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

func (lc LibraryConfig) Resources() []resources.Resource {
	return []resources.Resource{lc.Atlas, lc.Image}
}
