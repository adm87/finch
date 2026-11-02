package tiled_test

import (
	"testing"

	"github.com/go-openapi/testify/require"

	"github.com/adm87/finch/pkg/errs"
	"github.com/adm87/finch/plugins/tiled"
)

func TestConfigValidate(t *testing.T) {
	require.NoError(t, tiled.DefaultConfig().Validate())

	valid := tiled.Config{Maps: map[string]tiled.MapConfig{
		"gym04": {TMX: "gym04.tmx"},
	}}
	require.NoError(t, valid.Validate())

	missing := tiled.Config{Maps: map[string]tiled.MapConfig{"gym04": {}}}
	err := missing.Validate()
	require.ErrorIs(t, err, errs.ErrArgument)
	require.ErrorContains(t, err, `map "gym04" needs a .tmx file`)
}
