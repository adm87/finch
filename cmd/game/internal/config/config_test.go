package config_test

import (
	"testing"
	"testing/fstest"

	"github.com/go-openapi/testify/require"

	"github.com/adm87/finch/cmd/game/internal/config"
	"github.com/adm87/finch/content"
	"github.com/adm87/finch/pkg/errs"
	"github.com/adm87/finch/plugins/aseprite"
	"github.com/adm87/finch/plugins/tiled"
)

func TestLoadTheGameConfigs(t *testing.T) {
	embedded, err := content.GetFileSystem(content.EmbeddedFS)
	require.NoError(t, err)

	cfg := config.Default()
	require.NoError(t, config.Load(embedded, content.GameConfig.String(), cfg))
	require.NoError(t, cfg.Application.Validate())
	require.NoError(t, cfg.Engine.Validate())
	require.NotEmpty(t, cfg.Bundles)

	var aseCfg aseprite.Config
	require.NoError(t, config.Load(embedded, content.AsepriteConfig.String(), &aseCfg))
	require.NoError(t, aseCfg.Validate())

	var tiledCfg tiled.Config
	require.NoError(t, config.Load(embedded, content.TiledConfig.String(), &tiledCfg))
	require.NoError(t, tiledCfg.Validate())
}

func TestLoadKeepsDefaultsForMissingKeys(t *testing.T) {
	fsys := fstest.MapFS{
		"game.yaml":  {Data: []byte("engine:\n  title: Test\n")},
		"empty.yaml": {Data: nil},
	}

	cfg := config.Default()
	require.NoError(t, config.Load(fsys, "game.yaml", cfg))

	want := config.Default()
	want.Engine.Title = "Test"
	require.Equal(t, want, cfg)

	cfg = config.Default()
	require.NoError(t, config.Load(fsys, "empty.yaml", cfg))
	require.Equal(t, config.Default(), cfg)
}

func TestLoadErrors(t *testing.T) {
	fsys := fstest.MapFS{
		"typo.yaml":   {Data: []byte("application:\n  screen:\n    widht: 100\n")},
		"broken.yaml": {Data: []byte("engine: [\n")},
	}

	require.ErrorIs(t, config.Load(fsys, "typo.yaml", config.Default()), errs.ErrFormat)
	require.ErrorIs(t, config.Load(fsys, "broken.yaml", config.Default()), errs.ErrFormat)
	require.ErrorContains(t, config.Load(fsys, "missing.yaml", config.Default()), `read "missing.yaml"`)
}
