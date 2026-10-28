package tiled_test

import (
	"io"
	"testing"

	"github.com/go-openapi/testify/require"

	"github.com/adm87/finch/pkg/application"
	"github.com/adm87/finch/pkg/errs"
	"github.com/adm87/finch/pkg/logger"
	"github.com/adm87/finch/plugins/tiled"
)

func TestConfigValidate(t *testing.T) {
	valid := tiled.Config{Maps: map[string]tiled.MapConfig{
		"gym04": {FileSystem: "assets", TMX: "tiled/gym04.tmx"},
	}}
	require.NoError(t, valid.Validate())

	missing := tiled.Config{Maps: map[string]tiled.MapConfig{"gym04": {}}}
	err := missing.Validate()
	require.ErrorIs(t, err, errs.ErrArgument)
	require.ErrorContains(t, err, `map "gym04" needs a file system`)
	require.ErrorContains(t, err, `map "gym04" needs a .tmx file`)
}

func TestPluginInitChecksTheConfig(t *testing.T) {
	require.ErrorIs(t, tiled.NewPlugin(nil).Init(nil), errs.ErrArgumentNil)

	bad := &tiled.Config{Maps: map[string]tiled.MapConfig{"gym04": {}}}
	require.ErrorIs(t, tiled.NewPlugin(bad).Init(nil), errs.ErrArgument)

	app, err := application.New(application.DefaultConfig(), logger.NewLogger(io.Discard, logger.Options{}))
	require.NoError(t, err)
	plugin := tiled.NewPlugin(&tiled.Config{})
	require.NoError(t, plugin.Init(app))
	require.NotNil(t, plugin.Maps())
}
