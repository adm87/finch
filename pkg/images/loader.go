package images

import (
	"bytes"
	_ "image/jpeg" // register decoders used by image.Decode
	_ "image/png"
	"slices"

	"github.com/hajimehoshi/ebiten/v2/ebitenutil"

	"github.com/adm87/finch/pkg/types/errs"
)

var imageTypes = []string{"png", "jpg", "jpeg"}

func ImageTypes() []string {
	return slices.Clone(imageTypes)
}

type ImageLoader struct {
	cache *ImageCache
}

func NewImageLoader(cache *ImageCache) *ImageLoader {
	return &ImageLoader{cache: cache}
}

func (l *ImageLoader) Import(data []byte) (uint64, error) {
	img, _, err := ebitenutil.NewImageFromReader(bytes.NewReader(data))
	if err != nil {
		return 0, errs.Format("image", "decode").Wrap(err)
	}
	return l.cache.Put(img), nil
}

func (l *ImageLoader) Delete(handle uint64) error {
	img, err := l.cache.Delete(handle)
	if err != nil {
		return err
	}
	img.Deallocate()
	return nil
}
