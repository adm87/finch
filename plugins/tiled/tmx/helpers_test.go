package tmx_test

import (
	"encoding/xml"
	"io/fs"
	"os"
	"testing"

	"github.com/go-openapi/testify/require"

	"github.com/adm87/finch/plugins/tiled/tmx"
)

var (
	testdata = os.DirFS("testdata")
	assets   = os.DirFS("../../../content/assets")
)

func decodeFile[T any](t *testing.T, fsys fs.FS, path string) T {
	t.Helper()
	src, err := fs.ReadFile(fsys, path)
	require.NoError(t, err)

	var v T
	require.NoError(t, xml.Unmarshal(src, &v))
	return v
}

func decode(t *testing.T, src string) (tmx.Map, error) {
	t.Helper()
	var m tmx.Map
	return m, xml.Unmarshal([]byte(src), &m)
}

func layerByName(t *testing.T, m tmx.Map, name string) tmx.Layer {
	t.Helper()
	for l := range m.Layers.All() {
		if l.Base().Name == name {

			return l
		}
	}
	t.Fatalf("no layer %q", name)
	return tmx.Layer{}
}
