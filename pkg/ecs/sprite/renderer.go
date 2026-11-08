package sprite

import (
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/yohamta/donburi"

	"github.com/adm87/finch/pkg/ecs/render"
	"github.com/adm87/finch/pkg/ecs/transform"
	"github.com/adm87/finch/pkg/geom"
)

// Images resolves sprite handles to loaded images.
type Images interface {
	Get(handle uint64) *ebiten.Image
}

// NewRenderAdapter queues a draw command for each entry that has a Sprite and a Transform.
func NewRenderAdapter(images Images) render.Adapter {
	return render.AdapterFunc(func(queue *render.Queue, entries []*donburi.Entry, view render.View) {
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

			if !view.Bounds.Intersects(bounds(&geo, w, h)) {
				continue
			}
			geo.Concat(view.Matrix)

			r := render.Component.Get(e)
			queue.Push(render.Command{
				Image: img,
				Options: ebiten.DrawImageOptions{
					GeoM:       geo,
					ColorScale: s.color,
				},
				Layer: r.Layer(),
				Depth: r.Depth(),
			})
		}
	})
}

// bounds is the world-space box around a w×h image placed by geo.
func bounds(geo *ebiten.GeoM, w, h float64) geom.Rect {
	x0, y0 := geo.Apply(0, 0)
	x1, y1 := geo.Apply(w, 0)
	x2, y2 := geo.Apply(0, h)
	x3, y3 := geo.Apply(w, h)

	minX, minY := min(x0, x1, x2, x3), min(y0, y1, y2, y3)
	maxX, maxY := max(x0, x1, x2, x3), max(y0, y1, y2, y3)
	return geom.Rect{X: minX, Y: minY, Width: maxX - minX, Height: maxY - minY}
}
