package internal

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"

	"gopkg.in/yaml.v3"

	"github.com/adm87/finch/cmd/game/internal/game"
	"github.com/adm87/finch/content/embedded"
	"github.com/adm87/finch/content/resources"
	"github.com/adm87/finch/pkg/errs"
	"github.com/adm87/finch/pkg/logger"
	"github.com/adm87/finch/plugins/aseprite"
	"github.com/adm87/finch/plugins/tiled"
)

// resourcesRoot is relative to where the game is started.
const resourcesRoot = "content/resources"

func Boot() error {
	l := logger.NewLogger(os.Stdout, loggerOptions).Named("finch")

	cfgs, err := loadConfigs(embedded.Configs)
	if err != nil {
		return err
	}

	g, err := game.New(cfgs, newFiles(resourcesRoot), l)
	if err != nil {
		return err
	}
	if err := g.Preload(cfgs.Game.Preload); err != nil {
		return err
	}
	if err := g.MakeTestScene(); err != nil {
		return err
	}
	return g.Run(&cfgs.Game.Engine, l)
}

func loadConfigs(fsys fs.FS) (game.Configs, error) {
	cfgs := game.Configs{Game: game.DefaultConfig(), Tiled: tiled.DefaultConfig(), Aseprite: aseprite.DefaultConfig()}
	if err := loadConfig(fsys, embedded.GameConfig, cfgs.Game); err != nil {
		return game.Configs{}, fmt.Errorf("load game config: %w", err)
	}
	if err := loadConfig(fsys, embedded.TiledConfig, cfgs.Tiled); err != nil {
		return game.Configs{}, fmt.Errorf("load tiled config: %w", err)
	}
	if err := loadConfig(fsys, embedded.AsepriteConfig, cfgs.Aseprite); err != nil {
		return game.Configs{}, fmt.Errorf("load aseprite config: %w", err)
	}
	return cfgs, nil
}

// loadConfig decodes the YAML file at path into cfg. Keys the file leaves out
// keep cfg's current values, and unknown keys are errors.
func loadConfig[T any](fsys fs.FS, path string, cfg *T) error {
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

func newFiles(root string) game.Files {
	res := resources.New(root)
	return game.Files{
		Images:   embedded.Images,
		Aseprite: res.Aseprite,
		Tiled:    res.Tiled,
	}
}
