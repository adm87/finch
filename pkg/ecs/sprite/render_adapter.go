package sprite

import (
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/yohamta/donburi"

	"github.com/adm87/finch/pkg/ecs/rendering"
	"github.com/adm87/finch/pkg/ecs/transform"
)

// Images resolves sprite handles to loaded images.
type Images interface {
	Get(handle uint64) *ebiten.Image
}

// NewRenderAdapter draws entities that have a Sprite and a Transform.
func NewRenderAdapter(images Images) rendering.AdapterFunc {
	if images == nil {
		panic("sprite: nil images")
	}
	return func(queue *rendering.Queue, entries []*donburi.Entry, view rendering.View) {
		for _, e := range entries {
			if !e.HasComponent(Component) || !e.HasComponent(transform.Component) {
				continue
			}

			s := Component.Get(e)

			img := images.Get(s.handle)
			if img == nil {
				continue // not loaded yet
			}

			size := img.Bounds().Size()
			w, h := float64(size.X), float64(size.Y)

			var geo ebiten.GeoM
			geo.Translate(-s.anchorX*w, -s.anchorY*h)
			geo.Concat(transform.Component.Get(e).Matrix())
			geo.Concat(view.Matrix)

			r := rendering.Component.Get(e)
			queue.Push(rendering.Command{
				Image: img,
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
