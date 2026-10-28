package transform

import (
	"github.com/yohamta/donburi"

	"github.com/adm87/finch/pkg/geom"
)

// Bounds is the area an entity occupies, relative to its transform. Animation
// can change it every frame, so it is not a collision shape.
type Bounds struct {
	rect    geom.Rect
	version uint64
}

var BoundsComponent = donburi.NewComponentType[Bounds]()

func GetBounds(entry *donburi.Entry) (*Bounds, bool) {
	if !entry.HasComponent(BoundsComponent) {
		return nil, false
	}

	return BoundsComponent.Get(entry), true
}

func (b *Bounds) Rect() geom.Rect {
	return b.rect
}

func (b *Bounds) SetRect(rect geom.Rect) {
	if b.rect == rect {
		return
	}
	b.rect = rect
	b.version++
}

// Version changes whenever the rect changes.
func (b *Bounds) Version() uint64 {
	return b.version
}

func (b *Bounds) WorldRect(t *Transform) geom.Rect {
	return geom.TransformRect(t.GeoM(), b.Rect())
}
