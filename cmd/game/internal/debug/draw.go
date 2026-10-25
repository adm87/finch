package debug

import (
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"

	"github.com/adm87/finch/pkg/geom"
)

func applyFloat32(m geom.Matrix, x, y float64) (float32, float32) {
	sx, sy := m.Apply(x, y)
	return float32(sx), float32(sy)
}

// strokeWorldRect outlines a world-space rect. It still lines up when the camera rotates.
func strokeWorldRect(target *ebiten.Image, viewMatrix geom.Matrix, rect geom.Rect, width float32, clr color.Color) {
	minX, minY := rect.Min()
	maxX, maxY := rect.Max()
	x0, y0 := applyFloat32(viewMatrix, minX, minY)
	x1, y1 := applyFloat32(viewMatrix, maxX, minY)
	x2, y2 := applyFloat32(viewMatrix, maxX, maxY)
	x3, y3 := applyFloat32(viewMatrix, minX, maxY)

	vector.StrokeLine(target, x0, y0, x1, y1, width, clr, false)
	vector.StrokeLine(target, x1, y1, x2, y2, width, clr, false)
	vector.StrokeLine(target, x2, y2, x3, y3, width, clr, false)
	vector.StrokeLine(target, x3, y3, x0, y0, width, clr, false)
}
