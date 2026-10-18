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

func (r Rect) ToImageRectangle() image.Rectangle {
	minX, minY := r.Min()
	maxX, maxY := r.Max()
	return image.Rect(int(minX), int(minY), int(maxX), int(maxY))
}
