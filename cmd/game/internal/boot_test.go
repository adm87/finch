package internal

import (
	"testing"

	"github.com/go-openapi/testify/require"

	"github.com/adm87/finch/cmd/game/internal/config"
	"github.com/adm87/finch/cmd/game/internal/game"
	"github.com/adm87/finch/content"
	"github.com/adm87/finch/pkg/logger"
)

// TestBootSequence runs boot up to the game loop: config, game, bundles, then the scene.
func TestBootSequence(t *testing.T) {
	t.Chdir("../../..") // content's assets folder is found from the repo root
	require.NoError(t, content.InitializeFileSystems())

	embedded, err := content.GetFileSystem(content.EmbeddedFS)
	require.NoError(t, err)

	cfg := config.Default()
	require.NoError(t, config.Load(embedded, content.GameConfig.String(), cfg))

	g, err := game.New(cfg, embedded, logger.N)
	require.NoError(t, err)
	require.NoError(t, g.LoadBundles(cfg.Bundles))
	require.NoError(t, g.MakeTestScene())
}
