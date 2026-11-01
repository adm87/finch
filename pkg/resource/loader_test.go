package resource_test

import (
	"io/fs"
	"testing"
	"testing/fstest"

	"github.com/go-openapi/testify/require"

	"github.com/adm87/finch/pkg/errs"
	"github.com/adm87/finch/pkg/resource"
)

func TestLoaderResolvesSources(t *testing.T) {
	disk := fstest.MapFS{"a.txt": {Data: []byte("disk")}}
	sources := map[resource.Source]fs.FS{"disk": disk}
	loader := resource.NewLoader(sources)
	delete(sources, "disk")

	fsys, err := loader.FS("disk")
	require.NoError(t, err)
	raw, err := fs.ReadFile(fsys, "a.txt")
	require.NoError(t, err)
	require.Equal(t, "disk", string(raw))

	_, err = loader.FS("cdn")
	require.True(t, errs.IsNotFound(err))
}

func TestKeyKeepsSourcesApart(t *testing.T) {
	require.NotEqual(t, resource.Key("disk", "a.png"), resource.Key("cdn", "a.png"))
}
