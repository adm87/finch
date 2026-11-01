package tiled

import (
	"github.com/adm87/finch/pkg/images"
	"github.com/adm87/finch/pkg/rendering"
)

// Renderer exposes the plugin's renderer to tests.
func (p *Plugin) Renderer() rendering.Adapter {
	return newTileMapRenderer(p)
}

// Images exposes the plugin's tile images to tests.
func (p *Plugin) Images() *images.Cache {
	return p.images
}
