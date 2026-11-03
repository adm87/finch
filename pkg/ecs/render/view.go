package render

import (
	"github.com/hajimehoshi/ebiten/v2"

	"github.com/adm87/finch/pkg/geom"
)

// View is what a camera sees.
type View struct {
	Matrix ebiten.GeoM // world to target
	Bounds geom.Rect   // world-space area to draw; anything outside is culled
}
