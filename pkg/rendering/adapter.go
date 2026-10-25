package rendering

import (
	"github.com/hajimehoshi/ebiten/v2"

	"github.com/adm87/finch/pkg/geom"
)

// AdapterID refers to an Adapter registered with a Pipeline.
type AdapterID uint64

// Adapter draws one kind of render candidate.
type Adapter interface {
	Render(target *ebiten.Image, candidate *Candidate, viewport geom.Rect, matrix geom.Matrix)
}
