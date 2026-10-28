package rendering

import (
	"github.com/hajimehoshi/ebiten/v2"

	"github.com/adm87/finch/pkg/geom"
)

// AdapterID refers to an Adapter registered with a Pipeline.
type AdapterID uint64

// Adapter draws one kind of render candidate.
type Adapter interface {
	// Jobs returns rendering jobs for the provided candidate.
	// One or many jobs may be appended to the provided jobs slice and returned.
	Jobs(candidate *Candidate, jobs []Job) []Job

	// Render applies a single job to the target image.
	Render(target *ebiten.Image, job *Job, viewport geom.Rect, matrix geom.Matrix)
}
