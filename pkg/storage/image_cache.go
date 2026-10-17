package storage

import (
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
		return nil, cacheError("get", "image", handle, err)
	}
	return img, nil
}

func (ic *ImageCache) Set(handle uint64, img *ebiten.Image) error {
	if err := ic.store.Set(slotmap.Unpack(handle), img); err != nil {
		return cacheError("set", "image", handle, err)
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
		return nil, cacheError("delete", "image", handle, err)
	}
	return img, nil
}
