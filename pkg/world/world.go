package world

import (
	"errors"
	"fmt"

	"github.com/yohamta/donburi"

	"github.com/adm87/finch/pkg/geom"
	"github.com/adm87/finch/pkg/transform"
)

var baseArchetype = []donburi.IComponentType{
	transform.TransformComponent,
	transform.BoundsComponent,
	indexingComponent,
}

func aggregateComponentTypes(comps ...donburi.IComponentType) []donburi.IComponentType {
	if len(comps) == 0 {
		return baseArchetype
	}

	out := make([]donburi.IComponentType, 0, len(comps)+len(baseArchetype))
	out = append(out, baseArchetype...)
	out = append(out, comps...)

	return out
}

// World is an ECS world that indexes entities by their bounds for spatial queries.
type World struct {
	ecs        donburi.World
	partitions *partitions
}

func NewWorld(opts ...Option) *World {
	options := DefaultOptions()
	for _, opt := range opts {
		opt(options)
	}

	return &World{
		ecs:        donburi.NewWorld(),
		partitions: newPartitions(options.PartitionResolutions),
	}
}

func (w *World) ECS() donburi.World {
	return w.ecs
}

func (w *World) Create(comps ...donburi.IComponentType) *donburi.Entry {
	entry := w.ecs.Entry(w.ecs.Create(aggregateComponentTypes(comps...)...))

	return entry
}

func (w *World) CreateMany(count int, comps ...donburi.IComponentType) []*donburi.Entry {
	entries := make([]*donburi.Entry, count)
	comps = aggregateComponentTypes(comps...)

	for i := range entries {
		entries[i] = w.ecs.Entry(w.ecs.Create(comps...))
	}

	return entries
}

func (w *World) Destroy(entry *donburi.Entry) error {
	if !w.Valid(entry) {
		return nil
	}

	err := w.partitions.remove(entry)
	w.ecs.Remove(entry.Entity())
	return err
}

func (w *World) DestroyMany(entries []*donburi.Entry) error {
	var errs []error

	for _, entry := range entries {
		if !w.Valid(entry) {
			continue
		}

		if err := w.partitions.remove(entry); err != nil {
			errs = append(errs, fmt.Errorf("entity %v: %w", entry.Entity(), err))
		}

		w.ecs.Remove(entry.Entity())
	}

	if len(errs) > 0 {
		return fmt.Errorf("failed to destroy some entries: %w", errors.Join(errs...))
	}

	return nil
}

// Add indexes entry so queries can find it.
func (w *World) Add(entry *donburi.Entry) error {
	if !w.Valid(entry) {
		return nil
	}

	return w.partitions.add(entry)
}

func (w *World) AddMany(entries []*donburi.Entry) error {
	var errs []error

	for _, entry := range entries {
		if !w.Valid(entry) {
			continue
		}

		err := w.partitions.add(entry)
		if err != nil {
			errs = append(errs, fmt.Errorf("entity %v: %w", entry.Entity(), err))
		}
	}

	if len(errs) > 0 {
		return fmt.Errorf("failed to add some entries: %w", errors.Join(errs...))
	}

	return nil
}

// Update re-indexes entry after its transform or bounds change.
func (w *World) Update(entry *donburi.Entry) error {
	if !w.Valid(entry) {
		return nil
	}

	return w.partitions.update(entry)
}

func (w *World) UpdateMany(entries []*donburi.Entry) error {
	var errs []error

	for _, entry := range entries {
		if !w.Valid(entry) {
			continue
		}

		if err := w.partitions.update(entry); err != nil {
			errs = append(errs, fmt.Errorf("entity %v: %w", entry.Entity(), err))
		}
	}

	if len(errs) > 0 {
		return fmt.Errorf("failed to update some entries: %w", errors.Join(errs...))
	}

	return nil
}

func (w *World) Valid(entry *donburi.Entry) bool {
	if entry == nil {
		return false
	}
	return w.ecs.Valid(entry.Entity())
}

// Query appends the entities in region to dst. Pass dst[:0] to reuse a buffer.
func (w *World) Query(region geom.Rect, dst []*donburi.Entry) []*donburi.Entry {
	w.partitions.rangeOver(region, func(e donburi.Entity) bool {
		dst = append(dst, w.ecs.Entry(e))
		return true
	})
	return dst
}

// RangeOver calls fn for each entity in region until fn returns false.
func (w *World) RangeOver(region geom.Rect, fn func(*donburi.Entry) bool) {
	w.partitions.rangeOver(region, func(e donburi.Entity) bool {
		return fn(w.ecs.Entry(e))
	})
}
