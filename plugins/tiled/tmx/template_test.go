package tmx_test

import (
	"encoding/xml"
	"testing"

	"github.com/go-openapi/testify/require"

	"github.com/adm87/finch/pkg/errs"
	"github.com/adm87/finch/plugins/tiled/tmx"
)

func TestDecodesATemplate(t *testing.T) {
	tmpl := decodeFile[tmx.Template](t, testdata, "crate.tx")
	require.NoError(t, tmpl.Validate())
	require.Equal(t, "tilesets/props.tsx", tmpl.Tileset.Source)
	require.Equal(t, 1, tmpl.Tileset.FirstGID)
	require.Equal(t, "crate", tmpl.Object.Name)
	require.Equal(t, tmx.ShapeTile, tmpl.Object.Shape())
	require.True(t, tmpl.Object.Visible)
}

func TestTemplateValidate(t *testing.T) {
	embedded := tmx.Template{Tileset: &tmx.MapTileset{FirstGID: 1, Embedded: &tmx.Tileset{}}}
	require.ErrorIs(t, embedded.Validate(), errs.ErrNotSupported)

	noGID := tmx.Template{Tileset: &tmx.MapTileset{Source: "a.tsx"}}
	require.ErrorIs(t, noGID.Validate(), errs.ErrArgumentOutOfRange)

	badObject := tmx.Template{Object: tmx.Object{Width: -1}}
	require.ErrorIs(t, badObject.Validate(), errs.ErrArgumentOutOfRange)

	// A template without a tileset is a plain shape.
	require.NoError(t, tmx.Template{Object: tmx.Object{Opacity: 1}}.Validate())
}

func TestTemplateRejectsOtherRoots(t *testing.T) {
	var tmpl tmx.Template
	require.Error(t, xml.Unmarshal([]byte(`<object/>`), &tmpl))
}
