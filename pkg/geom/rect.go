package geom

import "image"

type Rect struct {
	X      float64 `json:"x"`
	Y      float64 `json:"y"`
	Width  float64 `json:"w"`
	Height float64 `json:"h"`
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

func (r Rect) Area() float64 {
	return r.Width * r.Height
}

// Empty reports whether the rect has no area.
func (r Rect) Empty() bool {
	return r.Width <= 0 || r.Height <= 0
}

// Union returns the smallest rect containing both r and other. An empty rect
// adds nothing, so a zero Rect can start a union.
func (r Rect) Union(other Rect) Rect {
	if other.Empty() {
		return r
	}
	if r.Empty() {
		return other
	}

	rMinX, rMinY := r.Min()
	rMaxX, rMaxY := r.Max()
	oMinX, oMinY := other.Min()
	oMaxX, oMaxY := other.Max()

	minX, minY := min(rMinX, oMinX), min(rMinY, oMinY)
	maxX, maxY := max(rMaxX, oMaxX), max(rMaxY, oMaxY)
	return Rect{X: minX, Y: minY, Width: maxX - minX, Height: maxY - minY}
}

func (r Rect) Intersects(other Rect) bool {
	rMinX, rMinY := r.Min()
	rMaxX, rMaxY := r.Max()
	oMinX, oMinY := other.Min()
	oMaxX, oMaxY := other.Max()
	return rMinX < oMaxX && rMaxX > oMinX && rMinY < oMaxY && rMaxY > oMinY
}

func (r Rect) Contains(other Rect) bool {
	rMinX, rMinY := r.Min()
	rMaxX, rMaxY := r.Max()
	oMinX, oMinY := other.Min()
	oMaxX, oMaxY := other.Max()
	return rMinX <= oMinX && rMaxX >= oMaxX && rMinY <= oMinY && rMaxY >= oMaxY
}

func (r Rect) ContainsPoint(x, y float64) bool {
	rMinX, rMinY := r.Min()
	rMaxX, rMaxY := r.Max()
	return x >= rMinX && x <= rMaxX && y >= rMinY && y <= rMaxY
}

func (r Rect) ClampPoint(x, y float64) (float64, float64) {
	minX, minY := r.Min()
	maxX, maxY := r.Max()
	return min(max(x, minX), maxX), min(max(y, minY), maxY)
}

func (r Rect) ClampRect(other Rect) Rect {
	other.X = clampSpan(other.X, other.Width, r.X, r.Width)
	other.Y = clampSpan(other.Y, other.Height, r.Y, r.Height)
	return other
}

func clampSpan(start, size, boundStart, boundSize float64) float64 {
	if size >= boundSize {
		return boundStart + (boundSize-size)/2
	}
	return min(max(start, boundStart), boundStart+boundSize-size)
}

func (r Rect) ToImageRectangle() image.Rectangle {
	minX, minY := r.Min()
	maxX, maxY := r.Max()
	return image.Rect(int(minX), int(minY), int(maxX), int(maxY))
}
