package transform

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

func GetRect(entry *donburi.Entry) geom.Rect {
	if !entry.HasComponent(BoundsComponent) {
		return geom.Rect{}
	}

	return BoundsComponent.Get(entry).rect
}

func SetRect(entry *donburi.Entry, rect geom.Rect) {
	if !entry.HasComponent(BoundsComponent) {
		entry.AddComponent(BoundsComponent)
		return
	}

	b := BoundsComponent.Get(entry)
	b.rect = rect
}

func GetWorldBounds(entry *donburi.Entry) geom.Rect {
	b, bOk := GetBounds(entry)
	if !bOk {
		return geom.Rect{}
	}

	t, tOk := GetTransform(entry)
	if !tOk {
		return b.rect
	}

	m, mOk := GetMatrix(entry)
	if !mOk {
		return b.rect
	}

	return GetWorldBoundsOf(b, t, m)
}

func GetWorldBoundsOf(b *Bounds, t *Transform, m *Matrix) geom.Rect {
	return TransformRect(t, m, b.rect)
}

func TransformRect(t *Transform, m *Matrix, rect geom.Rect) geom.Rect {
	x1, y1 := rect.Min()
	x2, y2 := rect.Max()

	geoM := GetGeoMOf(t, m)

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
