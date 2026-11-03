package render

import (
	"github.com/hajimehoshi/ebiten/v2"

	"github.com/adm87/finch/pkg/geom"
)

// View is the camera's view of the world.
type View struct {
	Matrix ebiten.GeoM // world space to target
	Bounds geom.Rect   // what the camera can see; everything else is culled
}
