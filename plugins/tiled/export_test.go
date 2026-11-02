package tiled

import "github.com/adm87/finch/pkg/rendering"

// Renderer exposes the plugin's renderer to tests.
func (p *Plugin) Renderer() rendering.Adapter {
	return newTileMapRenderer(p)
}
