package rendering

import (
	"github.com/hajimehoshi/ebiten/v2"

	"github.com/adm87/finch/pkg/geom"
)

// Adapter draws one kind of render candidate.
type Adapter interface {
	Render(target *ebiten.Image, candidate *Candidate, viewport geom.Rect, matrix ebiten.GeoM)
}
