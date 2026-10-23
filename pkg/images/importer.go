package images

import (
	"io"
	"slices"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"

	"github.com/adm87/finch/pkg/errs"
)

var imageTypes = []string{"png", "jpg", "jpeg"}

func Types() []string {
	return slices.Clone(imageTypes)
}

type Storage interface {
	Put(img *ebiten.Image) uint64
	Delete(handle uint64) (*ebiten.Image, error)
}

type Importer struct {
	store Storage
}

func NewImporter(store Storage) *Importer {
	return &Importer{
		store: store,
	}
}

func (l *Importer) Import(reader io.Reader) (uint64, error) {
	img, _, err := ebitenutil.NewImageFromReader(reader)
	if err != nil {
		return 0, errs.Format("image", "decode").Wrap(err)
	}
	return l.store.Put(img), nil
}

func (l *Importer) Delete(handle uint64) error {
	img, err := l.store.Delete(handle)
	if err != nil {
		return err
	}
	img.Deallocate()
	return nil
}
