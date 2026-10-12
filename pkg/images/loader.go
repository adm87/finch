package images

import (
	"bytes"
	"errors"
	"fmt"

	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
)

var (
	ErrImageDecode = errors.New("failed to decode image")
)

var imageTypes = []string{"png", "jpg", "jpeg"}

func ImageTypes() []string {
	return imageTypes
}

type ImageLoader struct {
	cache *ImageCache
}

func NewImageLoader(cache *ImageCache) *ImageLoader {
	return &ImageLoader{
		cache: cache,
	}
}

func (l *ImageLoader) Import(data []byte) (uint64, error) {
	img, _, err := ebitenutil.NewImageFromReader(bytes.NewBuffer(data))
	if err != nil {
		return 0, fmt.Errorf("%w: %v", ErrImageDecode, err)
	}
	return l.cache.Put(img), nil
}

func (l *ImageLoader) Delete(handle uint64) error {
	return nil
}
