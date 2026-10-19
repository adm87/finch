package rendering

import (
	"github.com/adm87/finch/pkg/geom"
	"github.com/hajimehoshi/ebiten/v2"
)

// Renderer defines an interface for rendering candidates onto a target image.
type Renderer interface {
	// Render renders the given candidate onto the target image.
	Render(target *ebiten.Image, candidate *RenderCandidate, viewport geom.Rect, matrix ebiten.GeoM)
}
