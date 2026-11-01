package tiled_test

import (
	"testing"

	"github.com/go-openapi/testify/require"

	"github.com/adm87/finch/pkg/assets"
	"github.com/adm87/finch/pkg/data"
	"github.com/adm87/finch/pkg/errs"
	"github.com/adm87/finch/pkg/images"
	"github.com/adm87/finch/pkg/logger"
	"github.com/adm87/finch/pkg/rendering"
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

func TestNewPluginChecksTheConfig(t *testing.T) {
	newPlugin := func(cfg *tiled.Config) (*tiled.Plugin, error) {
		return tiled.NewPlugin(cfg, assets.NewAssets(), data.NewCache(), images.NewCache(), rendering.NewPipeline(), logger.N)
	}

	_, err := newPlugin(nil)
	require.ErrorIs(t, err, errs.ErrArgumentNil)

	_, err = newPlugin(&tiled.Config{Maps: map[string]tiled.MapConfig{"gym04": {}}})
	require.ErrorIs(t, err, errs.ErrArgument)

	plugin, err := newPlugin(&tiled.Config{})
	require.NoError(t, err)
	require.NotNil(t, plugin.Maps())
}
