package images_test

import (
	"image"
	"testing"

	"github.com/go-openapi/testify/require"
	"github.com/hajimehoshi/ebiten/v2"

	"github.com/adm87/finch/pkg/errs"
	"github.com/adm87/finch/pkg/images"
)

func TestSliceFramesKeepsTrim(t *testing.T) {
	cache := images.NewCache()
	handle := cache.Put(ebiten.NewImage(64, 64))

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
	handle := cache.Put(ebiten.NewImage(64, 64))

	err := cache.SliceFrames(handle, []images.FrameRect{
		{Rect: image.Rect(0, 0, 24, 28), Offset: image.Pt(80, 0), Size: image.Pt(96, 96)},
	})
	require.ErrorIs(t, err, errs.ErrArgumentOutOfRange)
}

func TestFrameOfUnslicedImageIsUntrimmed(t *testing.T) {
	cache := images.NewCache()
	handle := cache.Put(ebiten.NewImage(10, 12))

	frame, err := cache.Frame(handle, 3)
	require.NoError(t, err)
	require.Equal(t, image.Point{}, frame.Offset)
	require.Equal(t, image.Pt(10, 12), frame.Size)
}
