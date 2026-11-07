package internal

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"io/fs"

	"go.yaml.in/yaml/v3"

	"github.com/adm87/finch/cmd/finch-game/internal/game"
	"github.com/adm87/finch/cmd/finch-game/internal/game/assets"
	"github.com/adm87/finch/cmd/finch-game/internal/game/draw"
	"github.com/adm87/finch/cmd/finch-game/internal/game/effects"
	"github.com/adm87/finch/cmd/finch-game/internal/game/scene"
	"github.com/adm87/finch/cmd/finch-game/internal/game/update"
	"github.com/adm87/finch/content/embedded"
	"github.com/adm87/finch/pkg/display"
	"github.com/adm87/finch/pkg/ecs"
	"github.com/adm87/finch/pkg/ecs/render"
	"github.com/adm87/finch/pkg/ecs/sprite"
	"github.com/adm87/finch/pkg/engine"
)

func Boot() error {
	config := game.DefaultConfig()
	if err := loadConfig(embedded.Configs, embedded.GameConfigPath, &config); err != nil {
		return err
	}
	return boot(config)
}

func boot(config game.Config) error {
	renderPipeline := render.NewPipeline()
	world := ecs.NewWorld(renderPipeline)
	screen := display.NewScreen(&config.Display.Screen)

	spriteKind := renderPipeline.Register(sprite.NewAdapter(assets.NewImages()))
	scene.LoadTest(world, spriteKind, screen.Visible(), 50)

	vignette, err := effects.NewVignette(1)
	if err != nil {
		return err
	}

	crt, err := effects.NewCRT(screen.Visible().Height)
	if err != nil {
		return err
	}

	drawer := draw.New(world, screen, vignette, crt)
	updater := update.New(world, vignette, crt)

	return engine.Run(
		engine.WithConfig(config.Engine),
		engine.WithDrawer(drawer.Pipeline()),
		engine.WithUpdater(updater.Main()),
		engine.WithLaterUpdater(updater.Later()),
		engine.WithDisplay(screen),
	)
}

func loadConfig(fsys fs.FS, path string, config any) error {
	data, err := fs.ReadFile(fsys, path)
	if err != nil {
		return err
	}
	dec := yaml.NewDecoder(bytes.NewReader(data))
	dec.KnownFields(true)
	if err := dec.Decode(config); err != nil && !errors.Is(err, io.EOF) {
		return fmt.Errorf("%s: %w", path, err)
	}
	return nil
}
