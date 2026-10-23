package data

import (
	"io"
	"slices"
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

func (dl *Importer) Import(reader io.Reader) (uint64, error) {
	raw, err := io.ReadAll(reader)
	if err != nil {
		return 0, err
	}
	return dl.cache.Put(raw), nil
}

func (dl *Importer) Delete(handle uint64) error {
	_, err := dl.cache.Delete(handle)
	return err
}
