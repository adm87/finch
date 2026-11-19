package sprite

import (
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/yohamta/donburi"

	"github.com/adm87/finch/pkg/assets"
	"github.com/adm87/finch/pkg/ecs/rendering"
	"github.com/adm87/finch/pkg/ecs/transform"
	"github.com/adm87/finch/pkg/images"
)

// Textures resolves sprite handles to loaded textures.
type Textures interface {
	Get(handle assets.Handle[*images.Texture]) (*images.Texture, bool)
}

// RenderAdapter draws entities that have a Sprite and a Transform.
func RenderAdapter(textures Textures) rendering.AdapterFunc {
	if textures == nil {
		panic("sprite: nil textures")
	}
	return func(queue *rendering.Queue, entries []*donburi.Entry, view rendering.View) {
		for _, e := range entries {
			if !e.HasComponent(Component) || !e.HasComponent(transform.Component) {
				continue
			}

			s := Component.Get(e)

			tex, ok := textures.Get(s.texture)
			if !ok {
				continue // not loaded yet
			}

			f := tex.Frame(s.frame)
			w, h := float64(f.Size.X), float64(f.Size.Y)

			var geo ebiten.GeoM
			geo.Translate(float64(f.Offset.X)-s.anchorX*w, float64(f.Offset.Y)-s.anchorY*h)
			geo.Concat(transform.Component.Get(e).Matrix())
			geo.Concat(view.Matrix)

			r := rendering.Component.Get(e)
			queue.Push(rendering.Command{
				Image: f.Image,
				Options: ebiten.DrawImageOptions{
					GeoM:       geo,
					ColorScale: s.color,
				},
				Layer: r.Layer(),
				Depth: r.Depth(),
			})
		}
	}
}
