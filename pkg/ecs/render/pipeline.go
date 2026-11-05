// Package render sorts entities by layer and depth and hands each one to the
// adapter that knows how to draw it.
package render

import (
	"slices"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/yohamta/donburi"
)

// Job is a single draw, sorted by layer and then depth.
type Job struct {
	Entry *donburi.Entry
	Layer int
	Depth int
	Index int // which part of the entry; up to the adapter
	kind  Kind
}

// Adapter draws one kind of entity.
type Adapter interface {
	// AppendJobs adds the entry's jobs that fall inside the view.
	// An entry can have any number of jobs.
	AppendJobs(e *donburi.Entry, view *View, jobs []Job) []Job

	// Draw draws one job.
	Draw(target *ebiten.Image, view *View, job *Job)
}

// Pipeline draws entities through the adapters registered with it.
type Pipeline struct {
	adapters []Adapter
	queue    queue
	view     View // lives here so taking its address doesn't allocate
}

func NewPipeline() *Pipeline {
	return &Pipeline{}
}

// Register adds an adapter and returns its Kind.
func (p *Pipeline) Register(a Adapter) Kind {
	if a == nil {
		panic("render: nil adapter")
	}
	if len(p.adapters) == maxKinds {
		panic("render: too many adapters")
	}
	p.adapters = append(p.adapters, a)
	return Kind(len(p.adapters))
}

// Draw sorts the visible entries by layer and depth, then draws them.
// Every entry needs a Renderer.
func (p *Pipeline) Draw(target *ebiten.Image, view View, entries []*donburi.Entry) {
	p.view = view
	q := &p.queue
	q.reset()

	for _, e := range entries {
		r := Component.Get(e)
		if !r.visible || r.kind == 0 {
			continue
		}

		start := len(q.jobs)
		q.jobs = p.adapters[r.kind-1].AppendJobs(e, &p.view, q.jobs)

		for i := start; i < len(q.jobs); i++ {
			j := &q.jobs[i]
			j.kind = r.kind
			q.keys = append(q.keys, key(j, i))
		}
	}

	slices.Sort(q.keys)

	for _, k := range q.keys {
		j := &q.jobs[k&indexMask]
		p.adapters[j.kind-1].Draw(target, &p.view, j)
	}
}
