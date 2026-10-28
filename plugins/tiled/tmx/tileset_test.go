package tmx_test

import (
	"encoding/xml"
	"testing"

	"github.com/go-openapi/testify/require"

	"github.com/adm87/finch/pkg/errs"
	"github.com/adm87/finch/plugins/tiled/tmx"
)

func TestDecodesTheRealTilesets(t *testing.T) {
	for _, name := range []string{"terrain", "trees"} {
		ts := decodeFile[tmx.Tileset](t, assets, "tiled/tilesets/"+name+".tsx")
		require.NoError(t, ts.Validate(), name)
		require.Equal(t, name, ts.Name)
		require.NotNil(t, ts.Image)
	}
}

func TestDecodesMapTilesets(t *testing.T) {
	m := decodeFile[tmx.Map](t, testdata, "full.tmx")
	require.Equal(t, "tilesets/terrain.tsx", m.Tilesets[0].Source)
	require.Equal(t, 1, m.Tilesets[0].FirstGID)
	require.Nil(t, m.Tilesets[0].Embedded)

	require.Equal(t, 49, m.Tilesets[1].FirstGID)
	require.Empty(t, m.Tilesets[1].Source)
	ts := m.Tilesets[1].Embedded
	require.NotNil(t, ts)

	require.Equal(t, "inline", ts.Name)
	require.Equal(t, 1, ts.Spacing)
	require.Equal(t, 2, ts.Margin)
	require.Equal(t, 4, ts.TileCount)
	require.Equal(t, 2, ts.Columns)
	require.Equal(t, tmx.AlignCenter, ts.ObjectAlignment)
	require.Equal(t, tmx.RenderSizeGrid, ts.TileRenderSize)
	require.Equal(t, tmx.FillPreserveAspectFit, ts.FillMode)
	require.Equal(t, "inline.png", ts.Image.Source)
	require.Equal(t, &tmx.Color{A: 0xff, R: 0xff, B: 0xff}, ts.Image.Trans)
	require.Equal(t, &tmx.TileOffset{X: 2, Y: -3}, ts.TileOffset)

	tile, ok := ts.Tile(1)
	require.True(t, ok)
	require.Equal(t, "water", tile.ClassName())
	require.Equal(t, []tmx.Frame{{TileID: 1, Duration: 100}, {TileID: 2, Duration: 150}}, tile.Animation)
	require.Equal(t, tmx.DrawIndex, tile.Collision.DrawOrder)
	require.Len(t, tile.Collision.Objects, 1)
	deep, _ := tile.Properties.Get("deep")
	require.Equal(t, "true", deep.Value)

	_, ok = ts.Tile(2)
	require.False(t, ok)
}

func TestDecodesACollectionTileset(t *testing.T) {
	ts := decodeFile[tmx.Tileset](t, testdata, "collection.tsx")
	require.NoError(t, ts.Validate())

	require.Equal(t, "decor", ts.Class)
	require.Nil(t, ts.Image)
	layer, _ := ts.Properties.Get("layer")
	require.Equal(t, "front", layer.Value)

	rock, ok := ts.Tile(0)
	require.True(t, ok)
	require.Equal(t, "rock", rock.ClassName())
	require.Equal(t, "rock.png", rock.Image.Source)

	sub, ok := ts.Tile(3)
	require.True(t, ok)
	require.Equal(t, 4, sub.X)
	require.Equal(t, 16, sub.Width)
	require.Equal(t, "png", sub.Image.Format)
	b, err := sub.Image.Data.Bytes()
	require.NoError(t, err)
	require.Equal(t, []byte("\x89PNG"), b)
}

func TestTilesetRejectsOtherRoots(t *testing.T) {
	var ts tmx.Tileset
	require.Error(t, xml.Unmarshal([]byte(`<map/>`), &ts))
}

func TestMapTilesetValidate(t *testing.T) {
	require.NoError(t, (&tmx.MapTileset{FirstGID: 1, Source: "a.tsx"}).Validate())
	require.ErrorIs(t, (&tmx.MapTileset{Source: "a.tsx"}).Validate(), errs.ErrArgumentOutOfRange)

	// An embedded tileset is checked too.
	bad := tmx.MapTileset{FirstGID: 1, Embedded: &tmx.Tileset{ObjectAlignment: "middle"}}
	require.ErrorIs(t, bad.Validate(), errs.ErrArgument)
}
