package tiled

import (
	"image"
	"testing"

	"github.com/go-openapi/testify/require"

	"github.com/adm87/finch/pkg/geom"
	"github.com/adm87/finch/plugins/tiled/tmx"
)

func TestFlipTileKeepsTheTileAtTheOrigin(t *testing.T) {
	const (
		flipH = tmx.GID(0x80000000)
		flipV = tmx.GID(0x40000000)
		flipD = tmx.GID(0x20000000)
	)
	// Where the top right corner of a 32×16 tile ends up.
	cases := map[string]struct {
		gid  tmx.GID
		x, y float64
	}{
		"none":       {1, 32, 0},
		"horizontal": {1 | flipH, 0, 0},
		"vertical":   {1 | flipV, 32, 16},
		"diagonal":   {1 | flipD, 0, 32},
		"all":        {1 | flipH | flipV | flipD, 16, 0},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			var m geom.Matrix
			flipTile(&m, tc.gid, 32, 16)
			x, y := m.Apply(32, 0)
			require.InDelta(t, tc.x, x, 1e-9)
			require.InDelta(t, tc.y, y, 1e-9)
		})
	}
}

func TestVisibleTilesPadsForOverdraw(t *testing.T) {
	mi := MapInfo{Map: &tmx.Map{TileWidth: 16, TileHeight: 16}}
	view := geom.Rect{X: 20, Y: 20, Width: 40, Height: 8}
	require.Equal(t, image.Rect(1, 1, 4, 2), mi.visibleTiles(view))

	mi.overdraw = image.Point{X: 16, Y: 32}
	require.Equal(t, image.Rect(0, -1, 5, 4), mi.visibleTiles(view))

	require.True(t, mi.visibleTiles(geom.Rect{}).Empty())
}

func TestLayerViewUndoesTheLayerTransform(t *testing.T) {
	var toWorld geom.Matrix
	toWorld.Translate(10, 5)
	toWorld.Scale(2, 2)

	got := layerView(toWorld, geom.Rect{X: 20, Y: 10, Width: 40, Height: 20})
	require.Equal(t, geom.Rect{X: 0, Y: 0, Width: 20, Height: 10}, got) // world is 2*(p+offset)

	var flat geom.Matrix
	flat.Scale(0, 0)
	require.Equal(t, geom.Rect{}, layerView(flat, geom.Rect{Width: 1, Height: 1}))
}
