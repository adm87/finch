package images_test

import (
	"bytes"
	"image"
	"image/png"
	"io/fs"
	"testing"
	"testing/fstest"

	"github.com/go-openapi/testify/require"

	"github.com/adm87/finch/pkg/images"
	"github.com/adm87/finch/pkg/resource"
)

func TestLoadStoresOnlyWhatIsMissing(t *testing.T) {
	var raw bytes.Buffer
	require.NoError(t, png.Encode(&raw, image.NewRGBA(image.Rect(0, 0, 3, 2))))
	files := fstest.MapFS{
		"sprite.png": {Data: raw.Bytes()},
		"broken.png": {Data: []byte("not a png")},
	}

	res := resource.New()
	require.NoError(t, res.Mount("disk", files))
	cache := images.NewCache()

	require.NoError(t, images.Load(res, cache, "disk", "sprite.png"))
	handle, ok := cache.Lookup(resource.Key("disk", "sprite.png"))
	require.True(t, ok)
	img, err := cache.Get(handle)
	require.NoError(t, err)
	require.Equal(t, image.Pt(3, 2), img.Bounds().Size())

	// Already cached, so the file isn't read again.
	delete(files, "sprite.png")
	require.NoError(t, images.Load(res, cache, "disk", "sprite.png"))

	err = images.Load(res, cache, "disk", "missing.png", "broken.png")
	require.ErrorIs(t, err, fs.ErrNotExist)
	require.ErrorContains(t, err, "decode disk:broken.png")
	_, ok = cache.Lookup(resource.Key("disk", "broken.png"))
	require.False(t, ok)
}
