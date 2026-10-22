package importers

import (
	"io"
	"slices"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"

	"github.com/adm87/finch/pkg/types/errs"
)

var imageTypes = []string{"png", "jpg", "jpeg"}

func ImageTypes() []string {
	return slices.Clone(imageTypes)
}

type ImageStorage interface {
	Put(img *ebiten.Image) uint64
	Delete(handle uint64) (*ebiten.Image, error)
}

type ImageLoader struct {
	store ImageStorage
}

func NewImageLoader(store ImageStorage) *ImageLoader {
	return &ImageLoader{
		store: store,
	}
}

func (l *ImageLoader) Import(reader io.Reader) (uint64, error) {
	img, _, err := ebitenutil.NewImageFromReader(reader)
	if err != nil {
		return 0, errs.Format("image", "decode").Wrap(err)
	}
	return l.store.Put(img), nil
}

func (l *ImageLoader) Delete(handle uint64) error {
	img, err := l.store.Delete(handle)
	if err != nil {
		return err
	}
	img.Deallocate()
	return nil
}
