// Package geom has the geometry types shared across the engine.
package geom

// Rect is an axis-aligned rectangle. Its max edges are exclusive.
type Rect struct {
	X      float64
	Y      float64
	Width  float64
	Height float64
}

func (r Rect) Min() (float64, float64) {
	return r.X, r.Y
}

func (r Rect) Max() (float64, float64) {
	return r.X + r.Width, r.Y + r.Height
}

func (r Rect) Intersects(other Rect) bool {
	rMinX, rMinY := r.Min()
	rMaxX, rMaxY := r.Max()
	oMinX, oMinY := other.Min()
	oMaxX, oMaxY := other.Max()
	return rMinX < oMaxX && rMaxX > oMinX && rMinY < oMaxY && rMaxY > oMinY
}
