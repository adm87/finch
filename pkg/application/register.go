package application

import (
	"github.com/adm87/finch/pkg/assets"
	"github.com/adm87/finch/pkg/data"
	"github.com/adm87/finch/pkg/images"
)

func registerAssetImporters(as *assets.Assets, c *Cache) error {
	if err := as.RegisterImporter(data.NewImporter(c.data), data.Types()); err != nil {
		return err
	}
	if err := as.RegisterImporter(images.NewImporter(c.images), images.Types()); err != nil {
		return err
	}
	return nil
}
