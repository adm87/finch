package internal

import (
	"testing"
	"testing/fstest"

	"github.com/go-openapi/testify/require"

	"github.com/adm87/finch/cmd/game/internal/game"
	"github.com/adm87/finch/content/embedded"
	"github.com/adm87/finch/pkg/errs"
	"github.com/adm87/finch/pkg/logger"
)

// TestBootSequence runs boot up to the game loop: configs, game, startup bundles, then the scene.
func TestBootSequence(t *testing.T) {
	cfgs, err := loadConfigs(embedded.Configs)
	require.NoError(t, err)

	res, err := newResources("../../../" + resourcesRoot)
	require.NoError(t, err)

	g, err := game.New(cfgs, res, logger.N)
	require.NoError(t, err)
	require.NoError(t, g.LoadBundles(cfgs.Game.Startup...))
	require.NoError(t, g.MakeTestScene())

	require.True(t, errs.IsNotFound(g.LoadBundles("nope")))
}

func TestLoadTheGameConfigs(t *testing.T) {
	cfgs, err := loadConfigs(embedded.Configs)
	require.NoError(t, err)
	require.NoError(t, cfgs.Game.Screen.Validate())
	require.NoError(t, cfgs.Game.Engine.Validate())
	require.NotEmpty(t, cfgs.Game.Startup)
	require.NoError(t, cfgs.Aseprite.Validate())
	require.NoError(t, cfgs.Tiled.Validate())
}

func TestLoadConfigKeepsDefaultsForMissingKeys(t *testing.T) {
	fsys := fstest.MapFS{
		"game.yaml":  {Data: []byte("engine:\n  title: Test\n")},
		"empty.yaml": {Data: nil},
	}

	cfg := game.DefaultConfig()
	require.NoError(t, loadConfig(fsys, "game.yaml", cfg))

	want := game.DefaultConfig()
	want.Engine.Title = "Test"
	require.Equal(t, want, cfg)

	cfg = game.DefaultConfig()
	require.NoError(t, loadConfig(fsys, "empty.yaml", cfg))
	require.Equal(t, game.DefaultConfig(), cfg)
}

func TestLoadConfigErrors(t *testing.T) {
	fsys := fstest.MapFS{
		"typo.yaml":   {Data: []byte("screen:\n  widht: 100\n")},
		"broken.yaml": {Data: []byte("engine: [\n")},
	}

	require.ErrorIs(t, loadConfig(fsys, "typo.yaml", game.DefaultConfig()), errs.ErrFormat)
	require.ErrorIs(t, loadConfig(fsys, "broken.yaml", game.DefaultConfig()), errs.ErrFormat)
	require.ErrorContains(t, loadConfig(fsys, "missing.yaml", game.DefaultConfig()), `read "missing.yaml"`)
}
