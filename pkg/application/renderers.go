package application

import (
	"github.com/adm87/finch/pkg/images"
	"github.com/adm87/finch/pkg/rendering"
)

type Renderers struct {
	Image rendering.AdapterID
}

func newRenderers(pipeline *rendering.Pipeline, cache *Cache) (*Renderers, error) {
	image, err := pipeline.RegisterRenderer(images.NewRenderer(cache.images))
	if err != nil {
		return nil, err
	}
	return &Renderers{
		Image: image,
	}, nil
}
