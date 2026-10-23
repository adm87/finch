package rendering

import (
	"github.com/hajimehoshi/ebiten/v2"

	"github.com/adm87/finch/pkg/geom"
)

// Adapter defines an interface for rendering candidates onto a target image.
type Adapter interface {
	// Render renders the given candidate onto the target image.
	Render(target *ebiten.Image, candidate *Candidate, viewport geom.Rect, matrix ebiten.GeoM)
}
