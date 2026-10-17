package images

import (
	"errors"
	"fmt"
	"strconv"

	"github.com/adm87/finch/pkg/types/errs"
	"github.com/adm87/finch/pkg/types/slotmap"
	"github.com/hajimehoshi/ebiten/v2"
)

type ImageCache struct {
	store *slotmap.SlotMap[*ebiten.Image]
}

func NewImageCache() *ImageCache {
	return &ImageCache{
		store: slotmap.New[*ebiten.Image](256),
	}
}

func (ic *ImageCache) Get(handle uint64) (*ebiten.Image, error) {
	img, err := ic.store.Get(slotmap.Unpack(handle))
	if err != nil {
		return nil, storeError("get", handle, err)
	}
	return img, nil
}

func (ic *ImageCache) Set(handle uint64, img *ebiten.Image) error {
	if err := ic.store.Set(slotmap.Unpack(handle), img); err != nil {
		return storeError("set", handle, err)
	}
	return nil
}

func (ic *ImageCache) Put(img *ebiten.Image) uint64 {
	k := ic.store.Insert(img)
	return k.Pack()
}

func (ic *ImageCache) Delete(handle uint64) (*ebiten.Image, error) {
	img, err := ic.store.Delete(slotmap.Unpack(handle))
	if err != nil {
		return nil, storeError("delete", handle, err)
	}
	return img, nil
}

func storeError(op string, handle uint64, err error) error {
	key := strconv.FormatUint(handle, 10)
	switch {
	case errors.Is(err, slotmap.ErrInvalidKey):
		return errs.Argument("handle", fmt.Sprintf("malformed image handle %s", key)).Wrap(err)
	case errors.Is(err, slotmap.ErrNotFound):
		return errs.NotFound("image", key).Wrap(err)
	default:
		return fmt.Errorf("%s image %s: %w", op, key, err)
	}
}
