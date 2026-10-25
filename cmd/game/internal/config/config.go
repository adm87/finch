package config

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"io/fs"

	"gopkg.in/yaml.v3"

	"github.com/adm87/finch/pkg/application"
	"github.com/adm87/finch/pkg/assets"
	"github.com/adm87/finch/pkg/engine"
	"github.com/adm87/finch/pkg/errs"
)

type GameConfig struct {
	Bundles     []assets.AssetBundle `yaml:"bundles"`
	Engine      engine.Config        `yaml:"engine"`
	Application application.Config   `yaml:"application"`
}

func Default() *GameConfig {
	return &GameConfig{
		Application: *application.DefaultConfig(),
		Engine:      *engine.DefaultConfig(),
	}
}

// Load decodes the YAML file at path into cfg. Keys the file leaves out keep
// cfg's current values, and unknown keys are errors.
func Load[T any](fsys fs.FS, path string, cfg *T) error {
	raw, err := fs.ReadFile(fsys, path)
	if err != nil {
		return fmt.Errorf("read %q: %w", path, err)
	}

	dec := yaml.NewDecoder(bytes.NewReader(raw))
	dec.KnownFields(true)
	if err := dec.Decode(cfg); err != nil && !errors.Is(err, io.EOF) {
		return errs.Format(path, "invalid yaml").Wrap(err)
	}
	return nil
}
