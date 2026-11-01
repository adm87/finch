package aseprite_test

import (
	"testing"
	"testing/fstest"

	"github.com/go-openapi/testify/require"

	"github.com/adm87/finch/pkg/errs"
	"github.com/adm87/finch/pkg/images"
	"github.com/adm87/finch/pkg/logger"
	"github.com/adm87/finch/plugins/aseprite"
)

func TestNewPluginChecksTheConfig(t *testing.T) {
	newPlugin := func(cfg *aseprite.Config) (*aseprite.Plugin, error) {
		return aseprite.NewPlugin(cfg, fstest.MapFS{}, images.NewCache(), logger.N)
	}

	_, err := newPlugin(nil)
	require.ErrorIs(t, err, errs.ErrArgumentNil)

	_, err = newPlugin(&aseprite.Config{Libraries: map[string]aseprite.LibraryConfig{"captain": {}}})
	require.ErrorIs(t, err, errs.ErrArgument)

	_, err = newPlugin(aseprite.DefaultConfig())
	require.NoError(t, err)
}
