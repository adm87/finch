package camera

import (
	"github.com/yohamta/donburi"

	"github.com/adm87/finch/pkg/transform"
	"github.com/adm87/finch/pkg/world"
)

var MainCamera = donburi.NewTag("main_camera")

// NewMain creates the main camera. It's left out of partitioning since
// nothing queries for it.
func NewMain(w *world.World) *donburi.Entry {
	return w.Create(MainCamera, world.DontPartition)
}

func Main(w *world.World) (*donburi.Entry, bool) {
	return MainCamera.First(w.ECS())
}

// Zoom is how much bigger the world looks through the camera. 2 shows
// everything at twice its size.
func Zoom(c *donburi.Entry) float64 {
	t, ok := transform.GetTransform(c)
	if !ok {
		return 1
	}
	sx, _ := t.Scale()
	return 1 / sx
}

func SetZoom(c *donburi.Entry, zoom float64) {
	t, ok := transform.GetTransform(c)
	if !ok || zoom <= 0 {
		return
	}
	t.SetScale(1/zoom, 1/zoom)
}

func ZoomBy(c *donburi.Entry, factor float64) {
	SetZoom(c, Zoom(c)*factor)
}
