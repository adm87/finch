package application

import (
	"github.com/adm87/finch/pkg/data"
	"github.com/adm87/finch/pkg/images"
)

type Cache struct {
	data   *data.Cache
	images *images.Cache
}

func NewCache() *Cache {
	return &Cache{
		data:   data.NewCache(),
		images: images.NewCache(),
	}
}

func (c *Cache) Data() *data.Cache {
	return c.data
}

func (c *Cache) Images() *images.Cache {
	return c.images
}
