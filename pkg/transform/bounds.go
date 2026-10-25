package transform

import (
	"github.com/yohamta/donburi"

	"github.com/adm87/finch/pkg/geom"
)

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
	rect := b.Rect()

	x1, y1 := rect.Min()
	x2, y2 := rect.Max()

	geoM := t.GeoM()

	ax, ay := geoM.Apply(x1, y1)
	bx, by := geoM.Apply(x2, y1)
	cx, cy := geoM.Apply(x2, y2)
	dx, dy := geoM.Apply(x1, y2)

	minX, maxX := min(ax, bx, cx, dx), max(ax, bx, cx, dx)
	minY, maxY := min(ay, by, cy, dy), max(ay, by, cy, dy)

	return geom.Rect{
		X:      minX,
		Y:      minY,
		Width:  maxX - minX,
		Height: maxY - minY,
	}
}
