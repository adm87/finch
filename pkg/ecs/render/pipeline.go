package render

import (
	"slices"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/yohamta/donburi"
)

// Adapter draws one kind of renderable entity.
type Adapter interface {
	// AppendJobs appends the entry's jobs inside the view to jobs and returns
	// the extended slice. An entry can have more than one job.
	AppendJobs(e *donburi.Entry, view *View, jobs []Job) []Job

	// Draw draws a single job to target through the view.
	Draw(target *ebiten.Image, view *View, job *Job)
}

// Job is one draw, ordered by layer, then depth.
type Job struct {
	Entry *donburi.Entry
	Layer int
	Depth int
	Index int // adapter-defined: which part of the entry
	kind  Kind
}

type Pipeline struct {
	adapters []Adapter
	queue    queue
	view     View // kept here so passing &view to adapters doesn't allocate
}

func NewPipeline() *Pipeline {
	return &Pipeline{}
}

// Register adds an adapter and returns the Kind that routes entities to it.
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

// Draw draws the visible entries through the view in layer and depth order.
// Every entry must have a Renderer.
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
