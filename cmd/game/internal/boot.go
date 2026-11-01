package internal

import (
	"fmt"
	"os"

	"github.com/adm87/finch/cmd/game/internal/config"
	"github.com/adm87/finch/cmd/game/internal/game"
	"github.com/adm87/finch/content"
	"github.com/adm87/finch/pkg/logger"
)

func Boot() error {
	l := logger.NewLogger(os.Stdout, loggerOptions).Named("finch")
	if err := content.InitializeFileSystems(); err != nil {
		return fmt.Errorf("setup file systems: %w", err)
	}

	embedded, err := content.GetFileSystem(content.EmbeddedFS)
	if err != nil {
		return err
	}

	cfg := config.Default()
	if err := config.Load(embedded, content.GameConfig.String(), cfg); err != nil {
		return fmt.Errorf("load game config: %w", err)
	}

	g, err := game.New(cfg, embedded, l)
	if err != nil {
		return err
	}
	if err := g.LoadBundles(cfg.Bundles); err != nil {
		return err
	}
	if err := g.MakeTestScene(); err != nil {
		return err
	}
	return g.Run(&cfg.Engine, l)
}
