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

func (r Rect) Center() (float64, float64) {
	return r.X + r.Width/2, r.Y + r.Height/2
}

func (r Rect) Size() (float64, float64) {
	return r.Width, r.Height
}

func (r Rect) Area() float64 {
	return r.Width * r.Height
}

func (r Rect) Contains(x, y float64) bool {
	minX, minY := r.Min()
	maxX, maxY := r.Max()
	return x >= minX && x < maxX && y >= minY && y < maxY
}

func (r Rect) Intersects(other Rect) bool {
	rMinX, rMinY := r.Min()
	rMaxX, rMaxY := r.Max()
	oMinX, oMinY := other.Min()
	oMaxX, oMaxY := other.Max()
	return rMinX < oMaxX && rMaxX > oMinX && rMinY < oMaxY && rMaxY > oMinY
}
