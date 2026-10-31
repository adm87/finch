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

// Import stores the image in the cache under the asset's path.
func (l *Importer) Import(asset assets.Asset, _ fs.FS, reader io.Reader) error {
	img, _, err := ebitenutil.NewImageFromReader(reader)
	if err != nil {
		return errs.Format("image", "decode").Wrap(err)
	}
	if _, err := l.cache.Put(asset.String(), img); err != nil {
		img.Deallocate()
		return err
	}
	return nil
}

// Delete removes the image from the cache and frees it.
func (l *Importer) Delete(asset assets.Asset) error {
	handle, exists := l.cache.Lookup(asset.String())
	if !exists {
		return errs.NotFound("image", asset.String())
	}
	img, err := l.cache.Delete(handle)
	if err != nil {
		return err
	}
	img.Deallocate()
	return nil
}

func (l *Importer) Lookup(name string) bool {
	_, exists := l.cache.Lookup(name)
	return exists
}
