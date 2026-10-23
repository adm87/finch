package importers

import (
	"io"
	"slices"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"

	"github.com/adm87/finch/pkg/errs"
)

var imageTypes = []string{"png", "jpg", "jpeg"}

func ImageTypes() []string {
	return slices.Clone(imageTypes)
}

type ImageStorage interface {
	Put(img *ebiten.Image) uint64
	Delete(handle uint64) (*ebiten.Image, error)
}

type ImageImporter struct {
	store ImageStorage
}

func NewImageLoader(store ImageStorage) *ImageImporter {
	return &ImageImporter{
		store: store,
	}
}

func (l *ImageImporter) Import(reader io.Reader) (uint64, error) {
	img, _, err := ebitenutil.NewImageFromReader(reader)
	if err != nil {
		return 0, errs.Format("image", "decode").Wrap(err)
	}
	return l.store.Put(img), nil
}

func (l *ImageImporter) Delete(handle uint64) error {
	img, err := l.store.Delete(handle)
	if err != nil {
		return err
	}
	img.Deallocate()
	return nil
}
