package world

import (
	"errors"
	"fmt"

	"github.com/yohamta/donburi"

	"github.com/adm87/finch/pkg/ecs/components"
	"github.com/adm87/finch/pkg/geom"
)

var baseArchetype = []donburi.IComponentType{
	components.TransformComponent,
	components.BoundsComponent,
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

// World represents an ECS world that manages entities and their components.
type World interface {
	// Create creates a new entity with the specified component types.
	Create(comps ...donburi.IComponentType) *donburi.Entry

	// CreateMany creates multiple new entities with the specified component types.
	CreateMany(count int, comps ...donburi.IComponentType) []*donburi.Entry

	// Destroy removes the specified entity from the world.
	Destroy(entry *donburi.Entry) error

	// DestroyMany removes multiple entities from the world.
	DestroyMany(entries []*donburi.Entry) error

	// Add adds the specified entity to the world.
	Add(entry *donburi.Entry) error

	// AddMany adds multiple entities to the world.
	AddMany(entries []*donburi.Entry) error

	// Update updates a single entity in the world.
	Update(entry *donburi.Entry) error

	// UpdateMany updates multiple entities in the world.
	UpdateMany(entries []*donburi.Entry) error

	// Query retrieves all entities within the specified region and appends them to the dst slice.
	// The dst slice is used to collect the results and can be reused to avoid allocations.
	Query(region geom.Rect, dst []*donburi.Entry) []*donburi.Entry

	// RangeOver iterates over all entities within the specified region and calls the yield function for each entity.
	// If the yield function returns false, the iteration stops immediately.
	RangeOver(region geom.Rect, fn func(*donburi.Entry) bool)

	// ECS returns the underlying donburi world.
	ECS() donburi.World
}

type wrld struct {
	ecs        donburi.World
	partitions *partitions
}

func NewWorld(opts ...WorldOption) World {
	options := DefaultWorldOptions()
	for _, opt := range opts {
		opt(options)
	}

	return &wrld{
		ecs:        donburi.NewWorld(),
		partitions: newPartitions(options.PartitionResolutions),
	}
}

func (w *wrld) ECS() donburi.World {
	return w.ecs
}

func (w *wrld) Create(comps ...donburi.IComponentType) *donburi.Entry {
	entry := w.ecs.Entry(w.ecs.Create(aggregateComponentTypes(comps...)...))

	return entry
}

func (w *wrld) CreateMany(count int, comps ...donburi.IComponentType) []*donburi.Entry {
	entries := make([]*donburi.Entry, count)
	comps = aggregateComponentTypes(comps...)

	for i := range entries {
		entries[i] = w.ecs.Entry(w.ecs.Create(comps...))
	}

	return entries
}

func (w *wrld) Destroy(entry *donburi.Entry) error {
	if !w.Valid(entry) {
		return nil
	}

	err := w.partitions.remove(entry)
	w.ecs.Remove(entry.Entity())
	return err
}

func (w *wrld) DestroyMany(entries []*donburi.Entry) error {
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

func (w *wrld) Add(entry *donburi.Entry) error {
	if !w.Valid(entry) {
		return nil
	}

	return w.partitions.add(entry)
}

func (w *wrld) AddMany(entries []*donburi.Entry) error {
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

func (w *wrld) Update(entry *donburi.Entry) error {
	if !w.Valid(entry) {
		return nil
	}

	return w.partitions.update(entry)
}

func (w *wrld) UpdateMany(entries []*donburi.Entry) error {
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

func (w *wrld) Valid(entry *donburi.Entry) bool {
	if entry == nil {
		return false
	}
	return w.ecs.Valid(entry.Entity())
}

func (w *wrld) Query(region geom.Rect, dst []*donburi.Entry) []*donburi.Entry {
	w.partitions.rangeOver(region, func(e donburi.Entity) bool {
		dst = append(dst, w.ecs.Entry(e))
		return true
	})
	return dst
}

func (w *wrld) RangeOver(region geom.Rect, fn func(*donburi.Entry) bool) {
	w.partitions.rangeOver(region, func(e donburi.Entity) bool {
		return fn(w.ecs.Entry(e))
	})
}
