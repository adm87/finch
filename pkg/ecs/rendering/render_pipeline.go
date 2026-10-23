package rendering

import (
	"errors"
	"fmt"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/yohamta/donburi"

	"github.com/adm87/finch/pkg/ecs/components"
	"github.com/adm87/finch/pkg/geom"
	"github.com/adm87/finch/pkg/slotmap"
)

type View interface {
	Viewport() (geom.Rect, ebiten.GeoM)
}

type World interface {
	RangeOver(region geom.Rect, fn func(*donburi.Entry) bool)
}

// RenderPipeline manages the rendering process by maintaining a list of renderers and a render queue.
type RenderPipeline struct {
	view      View
	world     World
	renderers *slotmap.SlotMap[Renderer]
	queue     *RenderQueue
}

func NewRenderPipeline(viewport View, world World) *RenderPipeline {
	return &RenderPipeline{
		view:      viewport,
		world:     world,
		renderers: slotmap.New[Renderer](64),
		queue:     NewRenderQueue(),
	}
}

func (rp *RenderPipeline) RegisterRenderer(renderer Renderer) (uint64, error) {
	k := rp.renderers.Insert(renderer)
	// TODO: Should probably add support for checking if an item exists in a slotmap.
	return k.Pack(), nil
}

func (rp *RenderPipeline) Draw(target *ebiten.Image) error {
	rp.queue.Clear()

	if rp.view == nil {
		return nil
	}

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
		renderer, err := rp.renderers.Get(key)

		if err != nil {
			rc.Clear()
			errs = append(errs, err)
			continue
		}

		renderer.Render(target, rc, rect, geoM)
		rc.Clear()
	}

	if len(errs) > 0 {
		return fmt.Errorf("render pipeline errors: %w", errors.Join(errs...))
	}

	return nil
}

func (rp *RenderPipeline) buildRenderQueue(region geom.Rect) {
	rp.world.RangeOver(region, func(entry *donburi.Entry) bool {
		if !entry.HasComponent(components.RendererComponent) {
			return true
		}

		r := components.RendererComponent.Get(entry)
		if !r.Visible() {
			return true
		}

		t, tOk := components.GetTransform(entry)
		m, mOk := components.GetMatrix(entry)
		b, bOk := components.GetBounds(entry)

		if !tOk || !mOk || !bOk {
			return true
		}

		if b.Rect().Area() == 0 {
			return true
		}

		if !b.WorldRect(t, m).Intersects(region) {
			return true
		}

		rp.queue.Enqueue(RenderCandidate{
			Entry:     entry,
			Renderer:  r,
			Transform: t,
			Matrix:    m,
			Bounds:    b,
		})
		return true
	})
}
