package tiled_test

import (
	"testing"
	"testing/fstest"

	"github.com/go-openapi/testify/require"

	"github.com/adm87/finch/pkg/errs"
	"github.com/adm87/finch/pkg/logger"
	"github.com/adm87/finch/pkg/rendering"
	"github.com/adm87/finch/plugins/tiled"
)

func TestNewPluginChecksTheConfig(t *testing.T) {
	newPlugin := func(cfg *tiled.Config) (*tiled.Plugin, error) {
		return tiled.NewPlugin(cfg, fstest.MapFS{}, rendering.NewPipeline(), logger.N)
	}

	_, err := newPlugin(nil)
	require.ErrorIs(t, err, errs.ErrArgumentNil)

	_, err = newPlugin(&tiled.Config{Maps: map[string]tiled.MapConfig{"gym04": {}}})
	require.ErrorIs(t, err, errs.ErrArgument)

	_, err = newPlugin(tiled.DefaultConfig())
	require.NoError(t, err)
}
