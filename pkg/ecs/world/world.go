package world

import (
	"errors"
	"fmt"
	"iter"

	"github.com/adm87/finch/pkg/geom"
	"github.com/yohamta/donburi"
)

var (
	ErrCreateFailed           = errors.New("world: create failed")
	ErrCreateManyFailed       = errors.New("world: create many failed")
	ErrCreateManyInvalidCount = errors.New("world: create many invalid count")
	ErrDestroyFailed          = errors.New("world: destroy failed")
	ErrUpdateFailed           = errors.New("world: update failed")
)

// World represents the ECS world interface that provides methods to interact with the
// underlying donburi world and manage WorldObjects.
type World interface {
	// ECS returns the underlying donburi.World of the ECS world.
	ECS() donburi.World

	// Create creates a new WorldObject in the ECS world with the specified components.
	Create(comps ...donburi.IComponentType) (WorldObject, error)

	// CreateMany creates multiple WorldObjects in the ECS world with the specified components.
	CreateMany(count int, comps ...donburi.IComponentType) ([]WorldObject, error)

	// Destroy removes the specified WorldObjects from the ECS world.
	Destroy(objs ...WorldObject) error

	// Update updates the specified WorldObjects in the ECS world, including partition state.
	// Call this after bounds/transform changes and after toggling tags that affect indexing
	// (for example adding/removing Unpartitioned).
	Update(objs ...WorldObject) error

	// Query returns objects whose world bounds intersect region.
	// Pass dst[:0] to reuse memory across calls.
	Query(region geom.Rect, dst []WorldObject) []WorldObject

	// Range returns an iterator over all world objects within the specified region.
	Range(region geom.Rect) iter.Seq[WorldObject]
}

type wrld struct {
	world      donburi.World
	partitions *partitions
	queryCache []WorldObject
}

func New(opts ...Option) World {
	options := DefaultOptions()
	for _, o := range opts {
		o(options)
	}
	return &wrld{
		world:      donburi.NewWorld(),
		partitions: makePartitions(options.Resolutions),
	}
}

func (w *wrld) ECS() donburi.World {
	return w.world
}

func (w *wrld) Create(comps ...donburi.IComponentType) (WorldObject, error) {
	obj := NewWorldObject(w.world, comps...)

	if err := w.partitions.add(w.world, obj); err != nil {
		w.world.Remove(obj.Entity())

		return 0, fmt.Errorf("%w: add object to partitions: %w", ErrCreateFailed, err)
	}

	return obj, nil
}

func (w *wrld) CreateMany(count int, comps ...donburi.IComponentType) ([]WorldObject, error) {
	if count < 0 {
		return nil, fmt.Errorf("%w: got %d", ErrCreateManyInvalidCount, count)
	}

	objs := make([]WorldObject, count)
	var err error

	for i := range objs {
		objs[i], err = w.Create(comps...)

		if err != nil {
			for j := range i {
				_ = w.partitions.remove(w.world, objs[j])
				w.world.Remove(objs[j].Entity())
			}

			return nil, fmt.Errorf("%w: create object %d/%d: %w", ErrCreateManyFailed, i+1, count, err)
		}
	}
	return objs, nil
}

func (w *wrld) Destroy(objs ...WorldObject) error {
	for _, obj := range objs {
		if err := w.partitions.remove(w.world, obj); err != nil {
			return fmt.Errorf("%w: remove object from partitions: %w", ErrDestroyFailed, err)
		}

		w.world.Remove(obj.Entity())
	}
	return nil
}

func (w *wrld) Update(objs ...WorldObject) error {
	for _, obj := range objs {
		if err := w.partitions.update(w.world, obj); err != nil {
			return fmt.Errorf("%w: update object in partitions: %w", ErrUpdateFailed, err)
		}
	}
	return nil
}

func (w *wrld) Query(region geom.Rect, dst []WorldObject) []WorldObject {
	w.queryCache = w.partitions.query(region, w.queryCache[:0])

	for _, obj := range w.queryCache {
		if !w.world.Valid(donburi.Entity(obj)) {
			continue
		}

		dst = append(dst, obj)
	}

	return dst
}

func (w *wrld) Range(region geom.Rect) iter.Seq[WorldObject] {
	return func(yield func(WorldObject) bool) {
		w.partitions.rangeOver(region, func(obj WorldObject) bool {
			if !w.world.Valid(donburi.Entity(obj)) {
				return true
			}

			return yield(obj)
		})
	}
}
