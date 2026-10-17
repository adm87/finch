package world

import (
	"errors"
	"slices"

	"github.com/adm87/finch/pkg/ecs/transform"
	"github.com/adm87/finch/pkg/geom"
	"github.com/adm87/finch/pkg/types/hashgrid"
	"github.com/adm87/finch/pkg/types/slotmap"
	"github.com/yohamta/donburi"
)

type indexing struct {
	key       hashgrid.Key
	partition int
}

var (
	indexingComponent = donburi.NewComponentType[indexing]()

	// DontPartition is a tag used to indicate that an entity should not be partitioned.
	DontPartition = donburi.NewTag("don't partition")
)

func getIndexing(entry *donburi.Entry) (*indexing, bool) {
	if !entry.HasComponent(indexingComponent) {
		return nil, false
	}
	idx := indexingComponent.Get(entry)
	if idx.partition == 0 && idx.key == 0 {
		idx.partition = -1
	}
	return idx, true
}

type partitions struct {
	grids []*hashgrid.HashGrid[donburi.Entity]
}

func newPartitions(resolutions []int) *partitions {
	if len(resolutions) == 0 {
		panic("no partition resolutions provided")
	}

	sorted := append([]int(nil), resolutions...)
	slices.Sort(sorted)

	grids := make([]*hashgrid.HashGrid[donburi.Entity], len(sorted))
	for i, res := range sorted {
		if res <= 0 {
			panic("resolutions must be greater than 0")
		}
		grids[i] = hashgrid.New[donburi.Entity](res, 256)
	}

	return &partitions{
		grids: grids,
	}
}

func (p *partitions) add(entry *donburi.Entry) error {
	idx, ok := getIndexing(entry)
	if !ok {
		return nil
	}
	if entry.HasComponent(DontPartition) {
		return p.delete(idx)
	}
	return p.reindex(entry, idx, p.getBounds(entry))
}

func (p *partitions) update(entry *donburi.Entry) error {
	return p.add(entry)
}

func (p *partitions) remove(entry *donburi.Entry) error {
	idx, ok := getIndexing(entry)
	if !ok {
		return nil
	}
	return p.delete(idx)
}

func (p *partitions) rangeOver(region geom.Rect, fn func(donburi.Entity) bool) {
	for _, g := range p.grids {
		for _, e := range g.Range(region) {
			if !fn(e) {
				return
			}
		}
	}
}

func (p *partitions) reindex(entry *donburi.Entry, idx *indexing, bounds geom.Rect) error {
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
	return p.insertInto(partition, entry, idx, bounds)
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

func (p *partitions) insertInto(partition int, entry *donburi.Entry, idx *indexing, bounds geom.Rect) error {
	key, err := p.grids[partition].Insert(entry.Entity(), bounds)
	if err != nil {
		return err
	}

	idx.partition, idx.key = partition, key
	return nil
}

func (p *partitions) getBounds(entry *donburi.Entry) geom.Rect {
	rect := transform.GetRect(entry)

	t, tOk := transform.GetTransform(entry)
	if !tOk {
		return rect
	}

	m, mOk := transform.GetMatrix(entry)
	if !mOk {
		return rect
	}

	return transform.TransformRect(t, m, rect)
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
