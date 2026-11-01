package tiled

import (
	"github.com/yohamta/donburi"

	"github.com/adm87/finch/pkg/rendering"
	"github.com/adm87/finch/pkg/transform"
	"github.com/adm87/finch/pkg/world"
)

// Spawn creates an entity that draws the map.
func (p *Plugin) Spawn(handle Handle, world *world.World) (*donburi.Entry, error) {
	info, err := p.Map(handle)
	if err != nil {
		return nil, err
	}

	e := world.Create(TileMapComponent, rendering.RendererComponent)

	b, _ := transform.GetBounds(e)
	b.SetRect(info.Bounds())

	r, _ := rendering.GetRenderer(e)
	r.SetRendererID(p.renderer)

	tm, _ := GetTileMap(e)
	tm.SetHandle(handle)

	return e, world.Update(e)
}
