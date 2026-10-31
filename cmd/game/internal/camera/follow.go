package camera

import (
	"github.com/yohamta/donburi"

	"github.com/adm87/finch/pkg/geom"
	"github.com/adm87/finch/pkg/transform"
)

func Follow(_ float64, c *donburi.Entry, target *donburi.Entry) {
	t, ok := transform.GetTransform(c)
	if !ok {
		return
	}
	tt, ok := transform.GetTransform(target)
	if !ok {
		return
	}

	tx, ty := tt.Position()
	t.SetPosition(tx, ty)
}

// Clamp moves the camera so view, its visible rect, stays inside bounds.
func Clamp(c *donburi.Entry, view, bounds geom.Rect) {
	t, ok := transform.GetTransform(c)
	if !ok {
		return
	}
	clamped := bounds.ClampRect(view)
	x, y := t.Position()
	t.SetPosition(x+clamped.X-view.X, y+clamped.Y-view.Y)
}
