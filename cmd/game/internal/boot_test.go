package internal

import (
	"testing"

	"github.com/go-openapi/testify/require"

	"github.com/adm87/finch/content"
	"github.com/adm87/finch/pkg/logger"
)

// TestBootSequence runs boot up to the game loop: app, plugins, then startup.
func TestBootSequence(t *testing.T) {
	t.Chdir("../../..") // content's assets folder is found from the repo root
	require.NoError(t, content.InitializeFileSystems())

	cfg, err := loadGameConfig()
	require.NoError(t, err)

	app, err := newApplication(cfg, logger.N)
	require.NoError(t, err)
	require.NoError(t, registerPlugins(app))

	require.NoError(t, loadBundles(cfg.Bundles)(app))
	require.NoError(t, makeTestScene(app))
}
