package images

import (
	"github.com/adm87/finch/pkg/resources"
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
	return img, resources.ConvertToResourceError(err)
}

func (ic *ImageCache) Set(handle uint64, img *ebiten.Image) error {
	err := ic.store.Set(slotmap.Unpack(handle), img)
	return resources.ConvertToResourceError(err)
}

func (ic *ImageCache) Put(img *ebiten.Image) uint64 {
	k := ic.store.Insert(img)
	return k.Pack()
}

func (ic *ImageCache) Delete(handle uint64) (*ebiten.Image, error) {
	img, err := ic.store.Delete(slotmap.Unpack(handle))
	return img, resources.ConvertToResourceError(err)
}
