package rendering

import (
	"errors"
	"fmt"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/yohamta/donburi"

	"github.com/adm87/finch/pkg/errs"
	"github.com/adm87/finch/pkg/geom"
	"github.com/adm87/finch/pkg/slotmap"
	"github.com/adm87/finch/pkg/transform"
)

// Pipeline draws the visible entities through their adapters, sorted by layer and depth.
type Pipeline struct {
	adapters  *slotmap.SlotMap[Adapter]
	queue     *Queue
	jobs      []Job     // scratch for one candidate's jobs
	candidate Candidate // reused, since adapters take it by pointer
}

func NewPipeline() *Pipeline {
	return &Pipeline{
		adapters: slotmap.New[Adapter](64),
		queue:    NewQueue(),
		jobs:     make([]Job, 0, 8),
	}
}

func (rp *Pipeline) RegisterRenderer(adapter Adapter) (AdapterID, error) {
	if adapter == nil {
		return 0, errs.ArgumentNil("adapter")
	}
	k := rp.adapters.Insert(adapter)
	// TODO: Should probably add support for checking if an item exists in a slotmap.
	return AdapterID(k.Pack()), nil
}

func (rp *Pipeline) Draw(target *ebiten.Image, entries []*donburi.Entry, viewRect geom.Rect, viewMatrix geom.Matrix) error {
	defer rp.queue.Clear()

	errList := rp.buildRenderQueue(entries, viewRect)
	for job := range rp.queue.Sorted() {
		job.adapter.Render(target, job, viewRect, viewMatrix)
	}

	if len(errList) > 0 {
		return fmt.Errorf("render pipeline errors: %w", errors.Join(errList...))
	}
	return nil
}

// buildRenderQueue asks each visible entity's adapter for its jobs and queues
// the ones in view.
func (rp *Pipeline) buildRenderQueue(entries []*donburi.Entry, viewRect geom.Rect) []error {
	var errList []error
	for _, entry := range entries {
		var ok bool
		if rp.candidate, ok = newCandidate(entry, viewRect); !ok {
			continue
		}

		adapter, err := rp.adapters.Get(slotmap.Unpack(uint64(rp.candidate.Renderer.RendererID())))
		if err != nil {
			errList = append(errList, err)
			continue
		}

		rp.jobs = adapter.Jobs(&rp.candidate, rp.jobs[:0])
		for _, job := range rp.jobs {
			if !job.Bounds.Intersects(viewRect) {
				continue
			}
			job.adapter = adapter
			if err := rp.queue.Enqueue(job); err != nil {
				return append(errList, err)
			}
		}
	}
	rp.candidate = Candidate{} // drop the last entry pointer
	return errList
}

// newCandidate skips entities that are hidden, have no bounds or are out of view.
func newCandidate(entry *donburi.Entry, viewRect geom.Rect) (Candidate, bool) {
	r, ok := GetRenderer(entry)
	if !ok || !r.Visible() {
		return Candidate{}, false
	}

	t, tOk := transform.GetTransform(entry)
	b, bOk := transform.GetBounds(entry)
	if !tOk || !bOk || b.Rect().Area() == 0 {
		return Candidate{}, false
	}

	bounds := b.WorldRect(t)
	if !bounds.Intersects(viewRect) {
		return Candidate{}, false
	}

	return Candidate{Entry: entry, Renderer: r, Transform: t, Bounds: bounds}, true
}
