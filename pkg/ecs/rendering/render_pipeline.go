package rendering

import (
	"errors"
	"fmt"

	"github.com/adm87/finch/pkg/ecs/components/renderer"
	"github.com/adm87/finch/pkg/ecs/components/transform"
	"github.com/adm87/finch/pkg/geom"
	"github.com/adm87/finch/pkg/types/slotmap"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/yohamta/donburi"
)

type View interface {
	Viewport() (geom.Rect, ebiten.GeoM)
}

type World interface {
	RangeOver(region geom.Rect, fn func(*donburi.Entry))
}

// RenderPipeline is responsible for managing render adapters and rendering entities within a specified viewport.
//
// Renderables must be partitioned within the World in order to be rendered, otherwise they will not appear in the render output.
type RenderPipeline struct {
	view     View
	world    World
	adapters *slotmap.SlotMap[RenderAdapter]
	queue    *RenderQueue
}

func NewRenderPipeline(viewport View, world World) *RenderPipeline {
	return &RenderPipeline{
		view:     viewport,
		world:    world,
		adapters: slotmap.New[RenderAdapter](64),
		queue:    NewRenderQueue(),
	}
}

func (rp *RenderPipeline) RegisterRenderer(adapter RenderAdapter) (uint64, error) {
	k := rp.adapters.Insert(adapter)
	// TODO: Should probably add support for checking if an item exists in a slotmap.
	return k.Pack(), nil
}

func (rp *RenderPipeline) Draw(target *ebiten.Image) error {
	rp.queue.Clear()

	rect, geoM := rp.view.Viewport()
	if rect.Area() == 0 {
		return nil
	}

	rp.buildRenderQueue(rect)

	var errs []error

	keys := rp.queue.SortedKeys()
	for i := range keys {
		idx := keys[i].Index()

		rc := rp.queue.CandidateAt(idx)
		if rc == nil || !rc.IsValid() {
			continue
		}

		key := slotmap.Unpack(rc.Renderer.ID())
		adapter, err := rp.adapters.Get(key)

		if err != nil {
			rc.Clear()
			errs = append(errs, err)
			continue
		}

		adapter.Render(target, rc, rect, geoM)
		rc.Clear()
	}

	if len(errs) > 0 {
		return fmt.Errorf("render pipeline errors: %w", errors.Join(errs...))
	}

	return nil
}

func (rp *RenderPipeline) buildRenderQueue(region geom.Rect) {
	rp.world.RangeOver(region, func(entry *donburi.Entry) {
		if !entry.HasComponent(renderer.RendererComponent) {
			return
		}

		r := renderer.RendererComponent.Get(entry)
		if !r.Visible() {
			return
		}

		t, tOk := transform.GetTransform(entry)
		m, mOk := transform.GetMatrix(entry)

		if !tOk || !mOk {
			return
		}

		bounds := transform.GetRect(entry)
		if bounds.Area() == 0 {
			return
		}

		if !transform.TransformRect(t, m, bounds).Intersects(region) {
			return
		}

		rp.queue.Enqueue(RenderCandidate{
			Entry:     entry,
			Renderer:  r,
			Transform: t,
			Matrix:    m,
		})
	})
}
