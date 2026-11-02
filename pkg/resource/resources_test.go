package resource_test

import (
	"io/fs"
	"testing"
	"testing/fstest"

	"github.com/go-openapi/testify/require"

	"github.com/adm87/finch/pkg/errs"
	"github.com/adm87/finch/pkg/resource"
)

func TestReadFileFromAMountedSource(t *testing.T) {
	res := resource.New()
	require.NoError(t, res.Mount("disk", fstest.MapFS{"a.txt": {Data: []byte("disk")}}))

	raw, err := res.ReadFile("disk", "a.txt")
	require.NoError(t, err)
	require.Equal(t, "disk", string(raw))

	_, err = res.ReadFile("disk", "missing.txt")
	require.ErrorIs(t, err, fs.ErrNotExist)
	require.ErrorContains(t, err, "disk:missing.txt")

	_, err = res.ReadFile("cdn", "a.txt")
	require.True(t, errs.IsNotFound(err))
}

func TestMountRejectsADuplicateSource(t *testing.T) {
	res := resource.New()
	require.NoError(t, res.Mount("disk", fstest.MapFS{}))
	require.True(t, errs.IsAlreadyExists(res.Mount("disk", fstest.MapFS{})))
}

func TestKeyKeepsSourcesApart(t *testing.T) {
	require.NotEqual(t, resource.Key("disk", "a.png"), resource.Key("cdn", "a.png"))
}
