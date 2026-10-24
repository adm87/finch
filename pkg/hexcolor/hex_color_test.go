package hexcolor_test

import (
	"encoding/json"
	"image/color"
	"testing"

	"github.com/go-openapi/testify/require"
	"gopkg.in/yaml.v3"

	"github.com/adm87/finch/pkg/hexcolor"
)

func TestParseHexColor(t *testing.T) {
	cases := []struct {
		in   string
		want hexcolor.Color
	}{
		{"#6495EDFF", hexcolor.Color{R: 0x64, G: 0x95, B: 0xED, A: 0xFF}},
		{"6495edff", hexcolor.Color{R: 0x64, G: 0x95, B: 0xED, A: 0xFF}},
		{"#00000000", hexcolor.Color{}},
		{"#fe5b5980", hexcolor.Color{R: 0xFE, G: 0x5B, B: 0x59, A: 0x80}},
	}

	for _, tc := range cases {
		var c hexcolor.Color
		require.NoError(t, c.ParseHexColor(tc.in), tc.in)
		require.Equal(t, tc.want, c, tc.in)
	}
}

func TestParseHexColorRejects(t *testing.T) {
	for _, in := range []string{"", "#", "#fff", "#ffffff", "#ffffffff00", "#gg0000ff"} {
		c := hexcolor.Color{R: 1, G: 2, B: 3, A: 4}
		require.Error(t, c.ParseHexColor(in), in)
		require.Equal(t, hexcolor.Color{R: 1, G: 2, B: 3, A: 4}, c, in)
	}
}

func TestToRGBA(t *testing.T) {
	c := hexcolor.Color{R: 1, G: 2, B: 3, A: 4}
	require.Equal(t, color.RGBA{R: 1, G: 2, B: 3, A: 4}, c.ToRGBA())
}

func TestUnmarshal(t *testing.T) {
	want := hexcolor.Color{R: 0x64, G: 0x95, B: 0xED, A: 0xFF}

	var fromJSON struct {
		Color hexcolor.Color `json:"color"`
	}
	require.NoError(t, json.Unmarshal([]byte(`{"color": "#6495EDFF"}`), &fromJSON))
	require.Equal(t, want, fromJSON.Color)

	var fromYAML struct {
		Color hexcolor.Color `yaml:"color"`
	}
	require.NoError(t, yaml.Unmarshal([]byte(`color: "#6495EDFF"`), &fromYAML))
	require.Equal(t, want, fromYAML.Color)

	require.Error(t, json.Unmarshal([]byte(`{"color": 42}`), &fromJSON))
	require.Error(t, yaml.Unmarshal([]byte(`color: "#fff"`), &fromYAML))
}
