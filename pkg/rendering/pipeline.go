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
	renderers *slotmap.SlotMap[Adapter]
	queue     *Queue
}

func NewPipeline() *Pipeline {
	return &Pipeline{
		renderers: slotmap.New[Adapter](64),
		queue:     NewQueue(),
	}
}

func (rp *Pipeline) RegisterRenderer(adapter Adapter) (AdapterID, error) {
	if adapter == nil {
		return 0, errs.ArgumentNil("adapter")
	}
	k := rp.renderers.Insert(adapter)
	// TODO: Should probably add support for checking if an item exists in a slotmap.
	return AdapterID(k.Pack()), nil
}

func (rp *Pipeline) Draw(target *ebiten.Image, entries []*donburi.Entry, viewRect geom.Rect, viewMatrix geom.Matrix) error {
	rp.buildRenderQueue(entries, viewRect)
	defer rp.queue.Clear()

	var errList []error

	keys := rp.queue.SortedKeys()
	for i := range keys {
		idx := keys[i].Index()

		rc := rp.queue.CandidateAt(idx)
		if rc == nil || !rc.IsValid() {
			continue
		}

		key := slotmap.Unpack(uint64(rc.Renderer.ID()))
		adapter, err := rp.renderers.Get(key)

		if err != nil {
			rc.Clear()
			errList = append(errList, err)
			continue
		}

		adapter.Render(target, rc, viewRect, viewMatrix)
		rc.Clear()
	}

	if len(errList) > 0 {
		return fmt.Errorf("render pipeline errors: %w", errors.Join(errList...))
	}

	return nil
}

func (rp *Pipeline) buildRenderQueue(entries []*donburi.Entry, viewRect geom.Rect) {
	for i := range entries {
		entry := entries[i]

		if !entry.HasComponent(RendererComponent) {
			continue
		}

		r := RendererComponent.Get(entry)
		if !r.Visible() {
			continue
		}

		t, tOk := transform.GetTransform(entry)
		b, bOk := transform.GetBounds(entry)

		if !tOk || !bOk {
			continue
		}

		if b.Rect().Area() == 0 {
			continue
		}

		if !b.WorldRect(t).Intersects(viewRect) {
			continue
		}

		rp.queue.Enqueue(Candidate{
			Entry:     entry,
			Renderer:  r,
			Transform: t,
			Bounds:    b,
		})
	}
}
