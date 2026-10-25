package images

import (
	"io"
	"io/fs"
	"slices"

	"github.com/hajimehoshi/ebiten/v2/ebitenutil"

	"github.com/adm87/finch/pkg/assets"
	"github.com/adm87/finch/pkg/errs"
)

var imageTypes = []string{"png", "jpg", "jpeg"}

func Types() []string {
	return slices.Clone(imageTypes)
}

type Importer struct {
	cache *Cache
}

func NewImporter(cache *Cache) *Importer {
	return &Importer{
		cache: cache,
	}
}

// Import returns an assets.Handle that converts to a Handle in the cache.
func (l *Importer) Import(_ assets.Asset, _ fs.FS, reader io.Reader) (assets.Handle, error) {
	img, _, err := ebitenutil.NewImageFromReader(reader)
	if err != nil {
		return 0, errs.Format("image", "decode").Wrap(err)
	}
	return assets.Handle(l.cache.Put(img)), nil
}

func (l *Importer) Delete(handle assets.Handle) error {
	img, err := l.cache.Delete(Handle(handle))
	if err != nil {
		return err
	}
	img.Deallocate()
	return nil
}
