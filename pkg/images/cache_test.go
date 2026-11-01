package images_test

import (
	"bytes"
	"image"
	"image/png"
	"io/fs"
	"testing"
	"testing/fstest"

	"github.com/go-openapi/testify/require"
	"github.com/hajimehoshi/ebiten/v2"

	"github.com/adm87/finch/pkg/errs"
	"github.com/adm87/finch/pkg/images"
)

func TestSliceFramesKeepsTrim(t *testing.T) {
	cache := images.NewCache()
	handle, err := cache.Put("img.png", ebiten.NewImage(64, 64))
	require.NoError(t, err)

	require.NoError(t, cache.SliceFrames(handle, []images.FrameRect{
		{Rect: image.Rect(0, 0, 24, 28), Offset: image.Pt(26, 36), Size: image.Pt(96, 96)},
		{Rect: image.Rect(24, 0, 40, 16)},
	}))

	trimmed, err := cache.Frame(handle, 0)
	require.NoError(t, err)
	require.Equal(t, image.Pt(24, 28), trimmed.Image.Bounds().Size())
	require.Equal(t, image.Pt(26, 36), trimmed.Offset)
	require.Equal(t, image.Pt(96, 96), trimmed.Size)

	plain, err := cache.Frame(handle, 1)
	require.NoError(t, err)
	require.Equal(t, image.Point{}, plain.Offset)
	require.Equal(t, image.Pt(16, 16), plain.Size)
}

func TestSliceFramesRejectsFramesOffTheirCanvas(t *testing.T) {
	cache := images.NewCache()
	handle, err := cache.Put("img.png", ebiten.NewImage(64, 64))
	require.NoError(t, err)

	err = cache.SliceFrames(handle, []images.FrameRect{
		{Rect: image.Rect(0, 0, 24, 28), Offset: image.Pt(80, 0), Size: image.Pt(96, 96)},
	})
	require.ErrorIs(t, err, errs.ErrArgumentOutOfRange)
}

func TestFrameOfUnslicedImageIsUntrimmed(t *testing.T) {
	cache := images.NewCache()
	handle, err := cache.Put("img.png", ebiten.NewImage(10, 12))
	require.NoError(t, err)

	frame, err := cache.Frame(handle, 3)
	require.NoError(t, err)
	require.Equal(t, image.Point{}, frame.Offset)
	require.Equal(t, image.Pt(10, 12), frame.Size)
}

func TestLoadDecodesOnceAndReusesTheHandle(t *testing.T) {
	var raw bytes.Buffer
	require.NoError(t, png.Encode(&raw, image.NewRGBA(image.Rect(0, 0, 3, 2))))
	fsys := fstest.MapFS{"sprite.png": {Data: raw.Bytes()}}

	cache := images.NewCache()
	handle, err := cache.Load("test:sprite.png", fsys, "sprite.png")
	require.NoError(t, err)
	img, err := cache.Get(handle)
	require.NoError(t, err)
	require.Equal(t, image.Pt(3, 2), img.Bounds().Size())

	again, err := cache.Load("test:sprite.png", fstest.MapFS{}, "sprite.png")
	require.NoError(t, err)
	require.Equal(t, handle, again)

	_, err = cache.Load("test:missing.png", fsys, "missing.png")
	require.ErrorIs(t, err, fs.ErrNotExist)
}
