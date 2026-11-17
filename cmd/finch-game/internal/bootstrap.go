package internal

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"io/fs"

	"go.yaml.in/yaml/v3"

	"github.com/adm87/finch/cmd/finch-game/internal/game"
	"github.com/adm87/finch/cmd/finch-game/internal/game/camera"
	"github.com/adm87/finch/cmd/finch-game/internal/game/draw"
	"github.com/adm87/finch/cmd/finch-game/internal/game/update"
	"github.com/adm87/finch/content/embedded"
	"github.com/adm87/finch/pkg/assets"
	"github.com/adm87/finch/pkg/display"
	"github.com/adm87/finch/pkg/ecs"
	"github.com/adm87/finch/pkg/ecs/rendering"
	"github.com/adm87/finch/pkg/ecs/sprite"
	"github.com/adm87/finch/pkg/engine"
	"github.com/adm87/finch/pkg/images"
)

func Boot() error {
	var content assets.Content
	if err := content.Mount(embedded.Assets); err != nil {
		return err
	}

	config := game.DefaultConfig()
	if err := loadConfig(&content, embedded.GameConfigPath, &config); err != nil {
		return err
	}

	return boot(config, &content)
}

func boot(config game.Config, fsys fs.FS) error {
	var router assets.Router

	imgs, err := images.Register(fsys, &router)
	if err != nil {
		return fmt.Errorf("register images: %w", err)
	}

	renderPipeline := rendering.NewPipeline(
		sprite.RenderAdapter(imgs),
	)

	world := ecs.NewWorld(renderPipeline)
	screen := display.NewScreen(config.Display.Screen)

	camera.New(world.Donburi(), 0, 0)

	drawer := draw.New(world, screen)
	updater := update.New(world, screen)

	return engine.Run(
		engine.WithConfig(config.Engine),
		engine.WithDrawer(drawer.Pipeline()),
		engine.WithUpdater(updater.Main()),
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
