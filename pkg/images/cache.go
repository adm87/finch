package images

import (
	"errors"
	"fmt"
	"image"
	"strconv"

	"github.com/hajimehoshi/ebiten/v2"

	"github.com/adm87/finch/pkg/errs"
	"github.com/adm87/finch/pkg/slotmap"
)

// Cache stores images and the frames they are sliced into.
type Cache struct {
	store  *slotmap.SlotMap[*ebiten.Image]
	frames map[uint64][]Frame
}

func NewCache() *Cache {
	return &Cache{
		store:  slotmap.New[*ebiten.Image](256),
		frames: make(map[uint64][]Frame),
	}
}

func (ic *Cache) Get(handle uint64) (*ebiten.Image, error) {
	img, err := ic.store.Get(slotmap.Unpack(handle))
	if err != nil {
		return nil, slotmap.HandleError("get", "image", handle, err)
	}
	return img, nil
}

func (ic *Cache) Put(img *ebiten.Image) uint64 {
	key := ic.store.Insert(img)
	return key.Pack()
}

// Set replaces the image at handle and clears its frames.
// The caller still owns the previous image.
func (ic *Cache) Set(handle uint64, img *ebiten.Image) error {
	if err := ic.store.Set(slotmap.Unpack(handle), img); err != nil {
		return slotmap.HandleError("set", "image", handle, err)
	}
	delete(ic.frames, handle)
	return nil
}

// Delete removes the image and its frames. The caller must deallocate the image.
func (ic *Cache) Delete(handle uint64) (*ebiten.Image, error) {
	img, err := ic.store.Delete(slotmap.Unpack(handle))
	if err != nil {
		return nil, slotmap.HandleError("delete", "image", handle, err)
	}
	delete(ic.frames, handle)
	return img, nil
}

// GetFrame returns frame index of the image. An image that was never sliced
// is treated as a single frame and returned for any index.
func (ic *Cache) GetFrame(handle uint64, index int) (Frame, error) {
	if frames, ok := ic.frames[handle]; ok {
		if index >= 0 && index < len(frames) {
			return frames[index], nil
		}
		return Frame{}, errs.ArgumentOutOfRange("index", index, fmt.Sprintf("image has %d frames", len(frames)))
	}
	img, err := ic.Get(handle)
	if err != nil {
		return Frame{}, err
	}
	return untrimmed(img), nil
}

// GetFrames returns the frames produced by the last slice call.
func (ic *Cache) GetFrames(handle uint64) ([]Frame, error) {
	frames, ok := ic.frames[handle]
	if !ok {
		return nil, errs.NotFound("image frames", strconv.FormatUint(handle, 10))
	}
	return frames, nil
}

// SliceUniform cuts the image into a grid of width×height frames, row by row.
// The image size must be an exact multiple of the frame size.
func (ic *Cache) SliceUniform(handle uint64, width, height int) error {
	img, err := ic.Get(handle)
	if err != nil {
		return err
	}

	b := img.Bounds()
	if err := errors.Join(
		errs.CheckRange("width", width, 1, b.Dx()),
		errs.CheckRange("height", height, 1, b.Dy()),
	); err != nil {
		return err
	}

	if b.Dx()%width != 0 || b.Dy()%height != 0 {
		msg := fmt.Sprintf("image %dx%d is not a multiple of frame %dx%d", b.Dx(), b.Dy(), width, height)
		return errs.Argument("width/height", msg)
	}

	rects := make([]FrameRect, 0, (b.Dx()/width)*(b.Dy()/height))
	for y := b.Min.Y; y < b.Max.Y; y += height {
		for x := b.Min.X; x < b.Max.X; x += width {
			rects = append(rects, FrameRect{Rect: image.Rect(x, y, x+width, y+height)})
		}
	}
	return ic.sliceFrames(img, handle, rects)
}

// SliceRects cuts the image into the given frames. Every rect must lie inside the image.
func (ic *Cache) SliceRects(handle uint64, rects []image.Rectangle) error {
	frameRects := make([]FrameRect, len(rects))
	for i, r := range rects {
		frameRects[i] = FrameRect{Rect: r}
	}
	return ic.SliceFrames(handle, frameRects)
}

// SliceFrames cuts the image into the given frames, which may be trimmed.
func (ic *Cache) SliceFrames(handle uint64, rects []FrameRect) error {
	img, err := ic.Get(handle)
	if err != nil {
		return err
	}
	return ic.sliceFrames(img, handle, rects)
}

func (ic *Cache) sliceFrames(img *ebiten.Image, handle uint64, rects []FrameRect) error {
	b := img.Bounds()
	frames := make([]Frame, len(rects))
	for i, fr := range rects {
		if fr.Rect.Empty() || !fr.Rect.In(b) {
			return errs.ArgumentOutOfRange("rects", fr.Rect, fmt.Sprintf("rect %d is empty or outside image bounds %v", i, b))
		}

		size := fr.Size
		if size == (image.Point{}) {
			size = fr.Rect.Size()
		}

		canvas := image.Rectangle{Max: size}
		if placed := fr.Rect.Sub(fr.Rect.Min).Add(fr.Offset); !placed.In(canvas) {
			return errs.ArgumentOutOfRange("rects", placed, fmt.Sprintf("rect %d does not fit on its %v canvas", i, size))
		}

		frames[i] = Frame{
			Image:  img.SubImage(fr.Rect).(*ebiten.Image),
			Offset: fr.Offset,
			Size:   size,
		}
	}
	ic.frames[handle] = frames

	return nil
}
