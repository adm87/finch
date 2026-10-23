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

// Cache is a cache for storing and slicing images into frames.
type Cache struct {
	store  *slotmap.SlotMap[*ebiten.Image]
	frames map[uint64][]*ebiten.Image
}

// NewCache creates and returns a new Cache instance.
func NewCache() *Cache {
	return &Cache{
		store:  slotmap.New[*ebiten.Image](256),
		frames: make(map[uint64][]*ebiten.Image),
	}
}

// Get retrieves the image associated with the given handle from the cache.
func (ic *Cache) Get(handle uint64) (*ebiten.Image, error) {
	img, err := ic.store.Get(slotmap.Unpack(handle))
	if err != nil {
		return nil, slotmap.HandleError("get", "image", handle, err)
	}
	return img, nil
}

// Put inserts the given image into the cache and returns its handle.
func (ic *Cache) Put(img *ebiten.Image) uint64 {
	key := ic.store.Insert(img)
	return key.Pack()
}

// Set replaces the image associated with the given handle with a new image.
// The previous image does not get deallocated, and should be managed by the caller.
//
// Frames associated associated with the previous image will be cleared.
func (ic *Cache) Set(handle uint64, img *ebiten.Image) error {
	if err := ic.store.Set(slotmap.Unpack(handle), img); err != nil {
		return slotmap.HandleError("set", "image", handle, err)
	}
	delete(ic.frames, handle)
	return nil
}

// Delete removes the image associated with the given handle from the cache.
//
// The image does not get deallocated, and should be managed by the caller.
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
func (ic *Cache) GetFrame(handle uint64, index int) (*ebiten.Image, error) {
	if frames, ok := ic.frames[handle]; ok {
		if index >= 0 && index < len(frames) {
			return frames[index], nil
		}
		return nil, errs.ArgumentOutOfRange("index", index, fmt.Sprintf("image has %d frames", len(frames)))
	}
	return ic.Get(handle)
}

// GetFrames returns the frames produced by the last SliceUniform or SliceRects call.
func (ic *Cache) GetFrames(handle uint64) ([]*ebiten.Image, error) {
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

	rects := make([]image.Rectangle, 0, (b.Dx()/width)*(b.Dy()/height))
	for y := b.Min.Y; y < b.Max.Y; y += height {
		for x := b.Min.X; x < b.Max.X; x += width {
			rects = append(rects, image.Rect(x, y, x+width, y+height))
		}
	}
	return ic.sliceRects(img, handle, rects)
}

// SliceRects cuts the image into the given frames. Every rect must lie inside the image.
func (ic *Cache) SliceRects(handle uint64, rects []image.Rectangle) error {
	img, err := ic.Get(handle)
	if err != nil {
		return err
	}
	return ic.sliceRects(img, handle, rects)
}

func (ic *Cache) sliceRects(img *ebiten.Image, handle uint64, rects []image.Rectangle) error {
	b := img.Bounds()
	for i, r := range rects {
		if r.Empty() || !r.In(b) {
			return errs.ArgumentOutOfRange("rects", r, fmt.Sprintf("rect %d is empty or outside image bounds %v", i, b))
		}
	}

	frames := make([]*ebiten.Image, len(rects))
	for i, r := range rects {
		frames[i] = img.SubImage(r).(*ebiten.Image)
	}
	ic.frames[handle] = frames

	return nil
}
