package display_test

import (
	"testing"

	"github.com/go-openapi/testify/require"

	"github.com/adm87/finch/pkg/display"
	"github.com/adm87/finch/pkg/errs"
)

func TestConfigValidate(t *testing.T) {
	require.NoError(t, display.DefaultConfig().Validate())

	cases := []struct {
		name          string
		width, height int
	}{
		{"zero size", 0, 0},
		{"zero width", 0, 600},
		{"negative height", 800, -1},
		{"larger than ebiten allows", 16385, 600},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			cfg := display.Config{Width: tc.width, Height: tc.height}
			require.ErrorIs(t, cfg.Validate(), errs.ErrArgumentOutOfRange)
		})
	}
}
