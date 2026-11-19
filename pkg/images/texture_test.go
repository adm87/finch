package images_test

import (
	"errors"
	"image"
	"testing"

	"github.com/hajimehoshi/ebiten/v2"

	"github.com/adm87/finch/pkg/images"
)

func TestNewTextureHasOneWholeFrame(t *testing.T) {
	tex := images.NewTexture(ebiten.NewImage(32, 16))

	if got := tex.Len(); got != 1 {
		t.Fatalf("Len() = %d, want 1", got)
	}
	f := tex.Frame(0)
	if f.Image != tex.Source() || f.Size != image.Pt(32, 16) {
		t.Errorf("Frame(0) = %v, want the whole 32x16 source", f)
	}
}

func TestSliceUniform(t *testing.T) {
	tex := images.NewTexture(ebiten.NewImage(40, 32))

	if err := tex.SliceUniform(16, 16); err != nil {
		t.Fatalf("SliceUniform(16, 16) = %v, want nil", err)
	}
	if got := tex.Len(); got != 4 {
		t.Fatalf("Len() = %d, want 4", got)
	}

	f := tex.Frame(1)
	if got, want := f.Image.Bounds(), image.Rect(16, 0, 32, 16); got != want {
		t.Errorf("Frame(1) bounds = %v, want %v", got, want)
	}
	if f.Size != image.Pt(16, 16) || f.Offset != (image.Point{}) {
		t.Errorf("Frame(1) = %v, %v, want size 16x16 and no offset", f.Size, f.Offset)
	}
}

func TestSliceRegionsKeepsTrim(t *testing.T) {
	tex := images.NewTexture(ebiten.NewImage(32, 32))

	err := tex.SliceRegions([]images.Region{
		{Rect: image.Rect(0, 0, 10, 12), Offset: image.Pt(3, 2), Size: image.Pt(16, 16)},
	})
	if err != nil {
		t.Fatalf("SliceRegions() = %v, want nil", err)
	}

	f := tex.Frame(0)
	if f.Offset != image.Pt(3, 2) || f.Size != image.Pt(16, 16) {
		t.Errorf("Frame(0) = %v, %v, want offset (3,2) and size 16x16", f.Offset, f.Size)
	}
}

func TestSliceFailsAndKeepsFrames(t *testing.T) {
	tests := []struct {
		slice func(*images.Texture) error
		name  string
	}{
		{name: "zero cell", slice: func(tex *images.Texture) error { return tex.SliceUniform(0, 16) }},
		{name: "cell larger than image", slice: func(tex *images.Texture) error { return tex.SliceUniform(64, 64) }},
		{name: "no rects", slice: func(tex *images.Texture) error { return tex.SliceRects(nil) }},
		{name: "rect outside image", slice: func(tex *images.Texture) error {
			return tex.SliceRects([]image.Rectangle{image.Rect(0, 0, 16, 16), image.Rect(24, 24, 40, 40)})
		}},
		{name: "empty rect", slice: func(tex *images.Texture) error {
			return tex.SliceRects([]image.Rectangle{image.Rect(4, 4, 4, 4)})
		}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tex := images.NewTexture(ebiten.NewImage(32, 32))

			if err := tt.slice(tex); !errors.Is(err, images.ErrInvalidSlice) {
				t.Errorf("slice = %v, want %v", err, images.ErrInvalidSlice)
			}
			if got := tex.Len(); got != 1 {
				t.Errorf("Len() = %d, want 1", got)
			}
		})
	}
}
