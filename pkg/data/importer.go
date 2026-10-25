package data

import (
	"io"
	"io/fs"
	"slices"

	"github.com/adm87/finch/pkg/assets"
)

var dataTypes = []string{"json", "yml", "yaml"}

func Types() []string {
	return slices.Clone(dataTypes)
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
func (dl *Importer) Import(_ assets.Asset, _ fs.FS, reader io.Reader) (assets.Handle, error) {
	raw, err := io.ReadAll(reader)
	if err != nil {
		return 0, err
	}
	return assets.Handle(dl.cache.Put(raw)), nil
}

func (dl *Importer) Delete(handle assets.Handle) error {
	_, err := dl.cache.Delete(Handle(handle))
	return err
}
