package internal

import (
	"testing"

	"github.com/go-openapi/testify/require"

	"github.com/adm87/finch/cmd/game/internal/plugins"
	"github.com/adm87/finch/content"
	"github.com/adm87/finch/pkg/logger"
)

func TestBootSetsUpApplicationAndPlugins(t *testing.T) {
	require.NoError(t, content.InitializeFileSystems())

	cfg, err := loadGameConfig()
	require.NoError(t, err)

	app, err := initializeApplication(cfg, logger.N)
	require.NoError(t, err)

	p, err := plugins.New(app, logger.N)
	require.NoError(t, err)
	require.NotNil(t, p.Aseprite)
}
