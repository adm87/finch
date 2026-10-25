package internal

import (
	"testing"

	"github.com/go-openapi/testify/require"

	"github.com/adm87/finch/cmd/game/internal/config"
	"github.com/adm87/finch/content"
	"github.com/adm87/finch/pkg/logger"
)

// TestBootSequence runs boot up to the game loop: config, app, plugins, then startup.
func TestBootSequence(t *testing.T) {
	t.Chdir("../../..") // content's assets folder is found from the repo root
	require.NoError(t, content.InitializeFileSystems())

	embedded, err := content.GetFileSystem(content.EmbeddedFS)
	require.NoError(t, err)

	cfg := config.Default()
	require.NoError(t, config.Load(embedded, content.GameConfig.String(), cfg))

	app, err := newApplication(cfg, logger.N)
	require.NoError(t, err)
	require.NoError(t, registerPlugins(app, embedded))

	require.NoError(t, loadBundles(cfg.Bundles)(app))
	require.NoError(t, makeTestScene(app))
}
