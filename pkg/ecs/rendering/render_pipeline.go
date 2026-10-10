package rendering

import (
	"errors"
	"iter"

	"github.com/adm87/finch/pkg/ecs/components/renderer"
	"github.com/adm87/finch/pkg/ecs/world"
	"github.com/adm87/finch/pkg/geom"
	"github.com/adm87/finch/pkg/types/slotmap"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/yohamta/donburi"
)

// Storage defines an interface for accessing world objects within a specific region.
type Storage interface {
	// Range returns an iterator over all world objects within the specified region.
	Range(region geom.Rect) iter.Seq[world.WorldObject]
}

// Viewport defines an interface for obtaining the current viewport and its associated transformation matrix.
type Viewport interface {
	// Get retrieves the current viewport rectangle and its associated transformation matrix.
	Get() (geom.Rect, ebiten.GeoM)
}

// RenderPipeline represents a rendering pipeline that uses a storage backend and a viewport to manage and render world objects.
type RenderPipeline struct {
	adapters *slotmap.SlotMap[RenderAdapter]
	queue    *RenderingQueue
	errs     []error
}

// NewRenderPipeline creates a new RenderPipeline with the given storage backend and viewport.
func NewRenderPipeline() *RenderPipeline {
	return &RenderPipeline{
		adapters: slotmap.New[RenderAdapter](128),
		queue:    NewRenderingQueue(),
	}
}

// AddAdapter adds a new render adapter to the pipeline and returns its unique identifier.
//
// Assign the returned adapter ID to a renderer component to link it with the appropriate render adapter.
func (rp *RenderPipeline) AddAdapter(adapter RenderAdapter) uint64 {
	k := rp.adapters.Insert(adapter)
	return k.Pack()
}

func (rp *RenderPipeline) Render(target *ebiten.Image, ecs donburi.World, storage Storage, viewport Viewport) error {
	rp.queue.Clear()

	region, geoM := viewport.Get()
	if region.Area() == 0 {
		return nil
	}

	for obj := range storage.Range(region) {
		entry := obj.Entry(ecs)

		if !entry.HasComponent(renderer.ComponentType) {
			continue
		}

		r := renderer.ComponentType.Get(entry)
		if !r.Visible() {
			continue
		}

		if !world.WorldBoundsOf(entry).Intersects(region) {
			continue
		}

		rp.queue.Enqueue(entry, r)
	}

	keys := rp.queue.SortedKeys()

	for i := range keys {
		idx := keys[i].Index()

		candidate := rp.queue.CandidateAt(idx)
		if !candidate.IsValid() {
			continue
		}

		key := slotmap.Unpack(candidate.Renderer.ID())
		adapter, err := rp.adapters.Get(key)

		if err != nil {
			rp.errs = append(rp.errs, err)
			continue
		}

		adapter.Render(target, candidate, region, geoM)
		candidate.Clear()
	}

	rp.queue.Clear()

	err := errors.Join(rp.errs...)
	rp.errs = rp.errs[:0]

	return err
}
