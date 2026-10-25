package internal

import (
	"testing"

	"github.com/go-openapi/testify/require"

	"github.com/adm87/finch/content"
	"github.com/adm87/finch/pkg/assets"
	"github.com/adm87/finch/pkg/logger"
	"github.com/adm87/finch/plugins/aseprite"
	"github.com/adm87/finch/plugins/tiled"
)

// TestBootSequence follows the order engine.Run uses: plugins, then startup.
func TestBootSequence(t *testing.T) {
	t.Chdir("../../..") // content's assets folder is found from the repo root
	require.NoError(t, content.InitializeFileSystems())

	cfg, err := loadGameConfig()
	require.NoError(t, err)

	app, err := newApplication(cfg, logger.N)
	require.NoError(t, err)

	aseCfg, err := loadConfig[aseprite.Config](app, assets.FileSystem(content.EmbeddedFS), content.AsepriteConfig)
	require.NoError(t, err)

	ase := aseprite.NewPlugin(aseCfg)
	require.NoError(t, ase.Init(app))

	tldCfg, err := loadConfig[tiled.Config](app, assets.FileSystem(content.EmbeddedFS), content.TiledConfig)
	require.NoError(t, err)

	tld := tiled.NewPlugin(tldCfg)
	require.NoError(t, tld.Init(app))

	require.NoError(t, app.RegisterPlugin(ase))
	require.NoError(t, app.RegisterPlugin(tld))

	require.NoError(t, loadBundles(cfg.Bundles)(app))
	require.NoError(t, makeTestScene(app))
}
