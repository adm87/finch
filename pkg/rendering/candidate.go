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

// Job is the candidate drawn once, which is all most adapters need.
func (rc *Candidate) Job() Job {
	return Job{
		Entry:     rc.Entry,
		Transform: rc.Transform,
		Bounds:    rc.Bounds,
		Layer:     rc.Renderer.Layer(),
		Depth:     rc.Renderer.Depth(),
	}
}
