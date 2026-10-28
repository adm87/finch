package rendering

import (
	"github.com/yohamta/donburi"

	"github.com/adm87/finch/pkg/geom"
	"github.com/adm87/finch/pkg/transform"
)

// Candidate is a visible entity waiting for its adapter to turn it into jobs.
type Candidate struct {
	Entry     *donburi.Entry
	Renderer  *Renderer
	Transform *transform.Transform
	Bounds    geom.Rect // world space
}

// Job returns a single rendering job for this candidate.
// Adapters may return multiple jobs for a candidate, but this is a convenience method for the common case of 'one job per candidate'.
func (rc *Candidate) Job() Job {
	return Job{
		Entry:     rc.Entry,
		Transform: rc.Transform,
		Bounds:    rc.Bounds,
		Layer:     rc.Renderer.Layer(),
		Depth:     rc.Renderer.Depth(),
	}
}
