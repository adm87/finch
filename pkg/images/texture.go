package images

import (
	"errors"
	"fmt"
	"image"

	"github.com/hajimehoshi/ebiten/v2"
)

var ErrInvalidSlice = errors.New("images: invalid slice")

// Frame is one region of a Texture. Offset and Size place Image inside its
// untrimmed frame, so trimmed frames still line up.
type Frame struct {
	Image  *ebiten.Image
	Offset image.Point
	Size   image.Point
}

// Region is where a frame is cut from the source, and how it was trimmed.
type Region struct {
	Rect   image.Rectangle
	Offset image.Point
	Size   image.Point
}

// Texture is a cached image and its frames. Frames are sub-images of the source.
type Texture struct {
	source *ebiten.Image
	frames []Frame
}

// NewTexture makes an unsliced texture with one frame, the whole image.
func NewTexture(source *ebiten.Image) *Texture {
	return &Texture{
		source: source,
		frames: []Frame{{Image: source, Size: source.Bounds().Size()}},
	}
}

func (t *Texture) Source() *ebiten.Image {
	return t.source
}

// Frame returns frame i, where 0 <= i < Len(). An unsliced texture has one frame, the whole image.
func (t *Texture) Frame(i int) Frame {
	return t.frames[i]
}

func (t *Texture) Len() int {
	return len(t.frames)
}

// SliceUniform cuts the source into w x h frames, left to right then top to bottom.
// Cells that don't fit at the right or bottom edge are dropped.
// Slicing an already sliced texture will replace the existing frames.
func (t *Texture) SliceUniform(w, h int) error {
	if w <= 0 || h <= 0 {
		return fmt.Errorf("slice %dx%d: %w", w, h, ErrInvalidSlice)
	}

	b := t.source.Bounds()
	var rects []image.Rectangle
	for y := b.Min.Y; y+h <= b.Max.Y; y += h {
		for x := b.Min.X; x+w <= b.Max.X; x += w {
			rects = append(rects, image.Rect(x, y, x+w, y+h))
		}
	}

	if len(rects) == 0 {
		return fmt.Errorf("slice %dx%d: larger than the %v image: %w", w, h, b.Size(), ErrInvalidSlice)
	}
	return t.SliceRects(rects)
}

// SliceRects cuts a frame for each rect, with no trimming.
// Slicing an already sliced texture will replace the existing frames.
func (t *Texture) SliceRects(rects []image.Rectangle) error {
	regions := make([]Region, len(rects))
	for i, r := range rects {
		regions[i] = Region{Rect: r, Size: r.Size()}
	}
	return t.SliceRegions(regions)
}

// SliceRegions cuts a frame for each region. If any region is empty or outside
// the source, it returns an error and keeps the existing frames.
// Slicing an already sliced texture will replace the existing frames.
func (t *Texture) SliceRegions(regions []Region) error {
	if len(regions) == 0 {
		return fmt.Errorf("slice: no regions: %w", ErrInvalidSlice)
	}

	b := t.source.Bounds()
	frames := make([]Frame, len(regions))
	for i, r := range regions {
		if r.Rect.Empty() || !r.Rect.In(b) {
			return fmt.Errorf("slice region %d %v: %w", i, r.Rect, ErrInvalidSlice)
		}
		frames[i] = Frame{
			Image:  t.source.SubImage(r.Rect).(*ebiten.Image),
			Offset: r.Offset,
			Size:   r.Size,
		}
	}

	t.frames = frames
	return nil
}
