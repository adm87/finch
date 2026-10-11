package geom

type Rect struct {
	X, Y          float64
	Width, Height float64
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
