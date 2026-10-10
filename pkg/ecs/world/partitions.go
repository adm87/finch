package world

import (
	"errors"
	"slices"

	"github.com/adm87/finch/pkg/geom"
	"github.com/adm87/finch/pkg/types/hashgrid"
	"github.com/adm87/finch/pkg/types/slotmap"
	"github.com/yohamta/donburi"
)

// Unpartitioned opts an object out of spatial partition indexing.
// Objects are partitioned by default. Attach this tag on creation (or later via game logic)
// for entities like cameras that should exist in the world but not occupy partition cells.
var Unpartitioned = donburi.NewTag("unpartitioned")

// indexing keeps track of an object's position within the spatial partitioning system.
type indexing struct {
	key       hashgrid.Key
	partition int
}

// index is a component that stores the spatial partitioning index for an object.
var index = donburi.NewComponentType[indexing](indexing{
	partition: -1,
})

// partitions manages multiple hashgrid partitions for spatial indexing of world objects.
type partitions struct {
	grids []*hashgrid.HashGrid[WorldObject]
}

func makePartitions(resolutions []int) *partitions {
	if len(resolutions) == 0 {
		panic("partitions: at least one resolution is required")
	}

	sorted := slices.Clone(resolutions)
	slices.Sort(sorted)
	if sorted[0] <= 0 {
		panic("partitions: resolutions must be positive")
	}

	p := &partitions{grids: make([]*hashgrid.HashGrid[WorldObject], len(sorted))}
	for i, res := range sorted {
		p.grids[i] = hashgrid.New[WorldObject](res, 256)
	}
	return p
}

func (p *partitions) add(world donburi.World, obj WorldObject) error {
	entry := obj.Entry(world)

	if entry.HasComponent(Unpartitioned) {
		return nil
	}

	return p.reindex(obj, index.Get(entry), p.getBounds(entry))
}

func (p *partitions) update(world donburi.World, obj WorldObject) error {
	return p.add(world, obj)
}

func (p *partitions) remove(world donburi.World, obj WorldObject) error {
	return p.delete(index.Get(obj.Entry(world)))
}

func (p *partitions) query(region geom.Rect, dst []WorldObject) []WorldObject {
	for _, g := range p.grids {
		dst = g.Query(region, dst)
	}
	return dst
}

func (p *partitions) rangeOver(region geom.Rect, yield func(WorldObject) bool) {
	for _, g := range p.grids {
		for _, obj := range g.Range(region) {
			if !yield(obj) {
				return
			}
		}
	}
}

func (p *partitions) reindex(obj WorldObject, idx *indexing, bounds geom.Rect) error {
	partition := p.nearestFit(bounds)

	if idx.partition == partition {
		err := p.grids[partition].Update(idx.key, bounds)
		if err == nil {
			return nil
		}
		if !errors.Is(err, slotmap.ErrNotFound) {
			return err
		}
		idx.partition, idx.key = -1, 0
	} else if err := p.delete(idx); err != nil {
		return err
	}

	// TODO(adm87): Cross-partition moves currently delete then insert. If insert fails,
	// the object can be dropped from indexing; revisit once hashgrid supports safer migration.
	return p.insertInto(partition, obj, idx, bounds)
}

func (p *partitions) delete(idx *indexing) error {
	if idx.partition < 0 {
		return nil
	}

	err := p.grids[idx.partition].Delete(idx.key)
	if err != nil && !errors.Is(err, slotmap.ErrNotFound) {
		return err
	}

	idx.partition, idx.key = -1, 0
	return nil
}

func (p *partitions) insertInto(partition int, obj WorldObject, idx *indexing, bounds geom.Rect) error {
	key, err := p.grids[partition].Insert(obj, bounds)
	if err != nil {
		return err
	}

	idx.partition, idx.key = partition, key
	return nil
}

func (p *partitions) getBounds(entry *donburi.Entry) geom.Rect {
	return WorldBoundsOf(entry)
}

func (p *partitions) nearestFit(bounds geom.Rect) int {
	size := max(bounds.Width, bounds.Height)

	for i, g := range p.grids {
		if float64(g.CellSize()) >= size {
			return i
		}
	}

	return len(p.grids) - 1
}
