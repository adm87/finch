package sprite

import (
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/yohamta/donburi"

	"github.com/adm87/finch/pkg/ecs/render"
	"github.com/adm87/finch/pkg/ecs/transform"
	"github.com/adm87/finch/pkg/geom"
)

// Images resolves a sprite handle to its image, or nil if there is none.
type Images interface {
	Image(handle uint64) *ebiten.Image
}

// Adapter draws entities that have a Sprite and a Transform.
type Adapter struct {
	images Images
	op     ebiten.DrawImageOptions
}

var _ render.Adapter = (*Adapter)(nil)

func NewAdapter(images Images) *Adapter {
	if images == nil {
		panic("sprite: nil images")
	}
	return &Adapter{images: images}
}

func (a *Adapter) AppendJobs(e *donburi.Entry, view *render.View, jobs []render.Job) []render.Job {
	s, ok := Get(e)
	if !ok || !e.HasComponent(transform.Component) {
		return jobs
	}
	img := a.images.Image(s.handle)
	if img == nil || !bounds(s, transform.Component.Get(e), img).Intersects(view.Bounds) {
		return jobs
	}

	r := render.Component.Get(e)
	return append(jobs, render.Job{Entry: e, Layer: r.Layer(), Depth: r.Depth()})
}

func (a *Adapter) Draw(target *ebiten.Image, view *render.View, job *render.Job) {
	s := Component.Get(job.Entry)
	t := transform.Component.Get(job.Entry)

	img := a.images.Image(s.handle)
	size := img.Bounds().Size()

	a.op.GeoM.Reset()
	a.op.GeoM.Translate(-s.anchorX*float64(size.X), -s.anchorY*float64(size.Y))
	a.op.GeoM.Concat(t.Matrix())
	a.op.GeoM.Concat(view.Matrix)
	a.op.ColorScale = s.color

	target.DrawImage(img, &a.op)
}

// bounds is the world-space box around the sprite's image after its transform.
func bounds(s *Sprite, t *transform.Transform, img *ebiten.Image) geom.Rect {
	size := img.Bounds().Size()
	w, h := float64(size.X), float64(size.Y)
	ox, oy := -s.anchorX*w, -s.anchorY*h

	m := t.Matrix()
	x0, y0 := m.Apply(ox, oy)
	x1, y1 := m.Apply(ox+w, oy)
	x2, y2 := m.Apply(ox, oy+h)
	x3, y3 := m.Apply(ox+w, oy+h)

	minX, minY := min(x0, x1, x2, x3), min(y0, y1, y2, y3)
	maxX, maxY := max(x0, x1, x2, x3), max(y0, y1, y2, y3)
	return geom.Rect{X: minX, Y: minY, Width: maxX - minX, Height: maxY - minY}
}
