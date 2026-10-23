package components

import (
	"github.com/adm87/finch/pkg/geom"
	"github.com/yohamta/donburi"
)

type Bounds struct {
	rect geom.Rect
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
	b.rect = rect
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
