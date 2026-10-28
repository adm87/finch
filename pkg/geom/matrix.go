package geom

import (
	"github.com/hajimehoshi/ebiten/v2"
)

type Matrix = ebiten.GeoM

// TransformRect returns the axis-aligned rect around r after m is applied.
func TransformRect(m Matrix, r Rect) Rect {
	x1, y1 := r.Min()
	x2, y2 := r.Max()

	ax, ay := m.Apply(x1, y1)
	bx, by := m.Apply(x2, y1)
	cx, cy := m.Apply(x2, y2)
	dx, dy := m.Apply(x1, y2)

	minX, maxX := min(ax, bx, cx, dx), max(ax, bx, cx, dx)
	minY, maxY := min(ay, by, cy, dy), max(ay, by, cy, dy)
	return Rect{X: minX, Y: minY, Width: maxX - minX, Height: maxY - minY}
}
