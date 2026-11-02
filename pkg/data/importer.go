package data

import (
	"io"
	"io/fs"
	"slices"

	"github.com/adm87/finch/pkg/assets"
	"github.com/adm87/finch/pkg/errs"
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

func (dl *Importer) Import(asset assets.Asset, _ fs.FS, reader io.Reader) error {
	raw, err := io.ReadAll(reader)
	if err != nil {
		return err
	}
	if _, err := dl.cache.Put(asset.String(), raw); err != nil {
		return err
	}
	return nil
}

func (dl *Importer) Delete(asset assets.Asset) error {
	handle, exists := dl.cache.Lookup(asset.String())
	if !exists {
		return errs.NotFound("data", asset.String())
	}
	if _, err := dl.cache.Delete(handle); err != nil {
		return err
	}
	return nil
}

func (dl *Importer) Lookup(name string) bool {
	_, exists := dl.cache.Lookup(name)
	return exists
}
