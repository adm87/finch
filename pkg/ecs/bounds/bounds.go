package bounds

import (
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/yohamta/donburi"

	"github.com/adm87/finch/pkg/geom"
)

type Bounds struct {
	local geom.Rect
}

var Component = donburi.NewComponentType[Bounds]()

func (b *Bounds) Local() geom.Rect {
	return b.local
}

func (b *Bounds) SetLocal(r geom.Rect) {
	b.local = r
}

func (b *Bounds) World(m ebiten.GeoM) geom.Rect {
	minX, minY := b.local.Min()
	maxX, maxY := b.local.Max()

	x0, y0 := m.Apply(minX, minY)
	x1, y1 := m.Apply(maxX, minY)
	x2, y2 := m.Apply(minX, maxY)
	x3, y3 := m.Apply(maxX, maxY)

	left, top := min(x0, x1, x2, x3), min(y0, y1, y2, y3)
	right, bottom := max(x0, x1, x2, x3), max(y0, y1, y2, y3)

	return geom.Rect{X: left, Y: top, Width: right - left, Height: bottom - top}
}
