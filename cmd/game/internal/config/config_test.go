package config_test

import (
	"io/fs"
	"testing"

	"github.com/go-openapi/testify/require"

	"github.com/adm87/finch/cmd/game/internal/config"
	"github.com/adm87/finch/content"
	"github.com/adm87/finch/pkg/application"
)

func TestParseGameConfigFile(t *testing.T) {
	fsys, err := content.GetFileSystem(content.EmbeddedFS)
	require.NoError(t, err)
	raw, err := fs.ReadFile(fsys, content.GameConfig.String())
	require.NoError(t, err)

	cfg, err := config.Parse(raw)
	require.NoError(t, err)
	require.NoError(t, cfg.Application.Validate())
	require.NotEmpty(t, cfg.Bundles)
}

func TestParseKeepsDefaultsForMissingKeys(t *testing.T) {
	cfg, err := config.Parse([]byte("application:\n  title: Test\n"))
	require.NoError(t, err)

	want := *application.DefaultConfig()
	want.Title = "Test"
	require.Equal(t, want, cfg.Application)
}

func TestParseRejectsUnknownKeys(t *testing.T) {
	_, err := config.Parse([]byte("application:\n  screen:\n    widht: 100\n"))
	require.Error(t, err)
}
