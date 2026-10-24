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

type View interface {
	Viewport() (geom.Rect, ebiten.GeoM)
}

type World interface {
	RangeOver(region geom.Rect, fn func(*donburi.Entry) bool)
}

// Pipeline draws the visible entities through their adapters, sorted by layer and depth.
type Pipeline struct {
	view      View
	world     World
	renderers *slotmap.SlotMap[Adapter]
	queue     *Queue
}

func NewPipeline(viewport View, world World) *Pipeline {
	return &Pipeline{
		view:      viewport,
		world:     world,
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

func (rp *Pipeline) Draw(target *ebiten.Image) error {
	rp.queue.Clear()

	if rp.view == nil {
		return nil
	}

	rect, geoM := rp.view.Viewport()
	if rect.Area() == 0 {
		return nil
	}

	rp.buildRenderQueue(rect)

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

		adapter.Render(target, rc, rect, geoM)
		rc.Clear()
	}

	if len(errList) > 0 {
		return fmt.Errorf("render pipeline errors: %w", errors.Join(errList...))
	}

	return nil
}

func (rp *Pipeline) buildRenderQueue(region geom.Rect) {
	rp.world.RangeOver(region, func(entry *donburi.Entry) bool {
		if !entry.HasComponent(RendererComponent) {
			return true
		}

		r := RendererComponent.Get(entry)
		if !r.Visible() {
			return true
		}

		t, tOk := transform.GetTransform(entry)
		b, bOk := transform.GetBounds(entry)

		if !tOk || !bOk {
			return true
		}

		if b.Rect().Area() == 0 {
			return true
		}

		if !b.WorldRect(t).Intersects(region) {
			return true
		}

		rp.queue.Enqueue(Candidate{
			Entry:     entry,
			Renderer:  r,
			Transform: t,
			Bounds:    b,
		})
		return true
	})
}
