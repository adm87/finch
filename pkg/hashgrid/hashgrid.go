package hashgrid

import (
	"iter"
	"math"

	"github.com/adm87/finch/pkg/geom"
	"github.com/adm87/finch/pkg/slotmap"
)

// Key finds an item in a Grid.
type Key slotmap.Key

// cellRange is a block of cells. The max edges aren't included.
type cellRange struct {
	minX, minY int32
	maxX, maxY int32
}

func (r cellRange) contains(x, y int32) bool {
	return x >= r.minX && x < r.maxX && y >= r.minY && y < r.maxY
}

// entry keeps the item's min cell so queries can skip repeats without a lookup.
type entry struct {
	key        Key
	minX, minY int32
}

type item[T any] struct {
	value T
	cells cellRange
}

// Grid files each item under every cell it touches.
type Grid[T any] struct {
	items   *slotmap.SlotMap[item[T]]
	cells   map[cell][]entry
	spare   [][]entry // emptied cell slices, for reuse
	size    float64
	invSize float64
	ranging int // active Range calls
}

// New panics if size isn't positive and finite.
func New[T any](size float64, capacity int) *Grid[T] {
	if !(size > 0) || math.IsInf(size, 1) {
		panic("hashgrid: size must be positive and finite")
	}
	return &Grid[T]{
		items:   slotmap.New[item[T]](capacity),
		cells:   make(map[cell][]entry, capacity),
		spare:   make([][]entry, 0, capacity),
		size:    size,
		invSize: 1 / size,
	}
}

func (g *Grid[T]) Len() int {
	return g.items.Len()
}

// Size is the width of a cell, in world units.
func (g *Grid[T]) Size() float64 {
	return g.size
}

// Insert always adds a new item, even if v is already in the grid.
// It panics on the same rects snap does.
func (g *Grid[T]) Insert(r geom.Rect, v T) Key {
	g.mustNotRange()

	cr := snap(r, g.invSize)
	sk := g.items.Insert(item[T]{value: v, cells: cr})

	k := Key(sk)
	e := entry{key: k, minX: cr.minX, minY: cr.minY}

	for y := cr.minY; y < cr.maxY; y++ {
		for x := cr.minX; x < cr.maxX; x++ {
			g.add(packCell(x, y), e)
		}
	}

	return k
}

// Update reports false if k was deleted. It panics on the same rects snap does.
func (g *Grid[T]) Update(k Key, r geom.Rect) bool {
	g.mustNotRange()

	it, ok := g.items.GetPtr(slotmap.Key(k))
	if !ok {
		return false
	}

	cr := snap(r, g.invSize)
	if cr == it.cells {
		return true
	}

	old := it.cells
	for y := old.minY; y < old.maxY; y++ {
		for x := old.minX; x < old.maxX; x++ {
			if !cr.contains(x, y) {
				g.remove(packCell(x, y), k)
			}
		}
	}

	e := entry{key: k, minX: cr.minX, minY: cr.minY}
	moved := old.minX != cr.minX || old.minY != cr.minY
	for y := cr.minY; y < cr.maxY; y++ {
		for x := cr.minX; x < cr.maxX; x++ {
			switch {
			case !old.contains(x, y):
				g.add(packCell(x, y), e)
			case moved:
				g.set(packCell(x, y), e)
			}
		}
	}

	it.cells = cr
	return true
}

// Delete reports false if k was already deleted.
func (g *Grid[T]) Delete(k Key) (T, bool) {
	g.mustNotRange()

	var zero T
	it, ok := g.items.Delete(slotmap.Key(k))
	if !ok {
		return zero, false
	}

	cr := it.cells
	for y := cr.minY; y < cr.maxY; y++ {
		for x := cr.minX; x < cr.maxX; x++ {
			g.remove(packCell(x, y), k)
		}
	}

	return it.value, true
}

// Query appends each item whose cells overlap r's. It's rough, so check
// exact bounds yourself.
func (g *Grid[T]) Query(r geom.Rect, dst []T) []T {
	q := snap(r, g.invSize)
	for y := q.minY; y < q.maxY; y++ {
		for x := q.minX; x < q.maxX; x++ {
			for _, e := range g.cells[packCell(x, y)] {
				if x != max(e.minX, q.minX) || y != max(e.minY, q.minY) {
					continue // its first cell in q reports it
				}
				it, _ := g.items.GetPtr(slotmap.Key(e.key)) // cells only hold live keys
				dst = append(dst, it.value)
			}
		}
	}
	return dst
}

// Range is Query with keys, as an iterator. Changing the grid inside it panics.
func (g *Grid[T]) Range(r geom.Rect) iter.Seq2[Key, T] {
	q := snap(r, g.invSize)
	return func(yield func(Key, T) bool) {
		g.ranging++
		defer g.endRange()

		for y := q.minY; y < q.maxY; y++ {
			for x := q.minX; x < q.maxX; x++ {
				for _, e := range g.cells[packCell(x, y)] {
					if x != max(e.minX, q.minX) || y != max(e.minY, q.minY) {
						continue
					}
					it, _ := g.items.GetPtr(slotmap.Key(e.key))
					if !yield(e.key, it.value) {
						return
					}
				}
			}
		}
	}
}

func (g *Grid[T]) add(c cell, e entry) {
	entries, ok := g.cells[c]

	if !ok {
		if n := len(g.spare); n > 0 {
			entries = g.spare[n-1]
			g.spare = g.spare[:n-1]
		}
	}

	g.cells[c] = append(entries, e)
}

func (g *Grid[T]) remove(c cell, k Key) {
	entries := g.cells[c]

	for i := range entries {
		if entries[i].key == k {
			last := len(entries) - 1
			entries[i] = entries[last]
			entries = entries[:last]
			break
		}
	}

	if len(entries) == 0 {
		delete(g.cells, c)
		g.spare = append(g.spare, entries)
		return
	}

	g.cells[c] = entries
}

// set updates e's min cell in c.
func (g *Grid[T]) set(c cell, e entry) {
	entries := g.cells[c]
	for i := range entries {
		if entries[i].key == e.key {
			entries[i] = e
			return
		}
	}
}

func (g *Grid[T]) endRange() {
	g.ranging--
}

func (g *Grid[T]) mustNotRange() {
	if g.ranging > 0 {
		panic("hashgrid: Insert, Update or Delete during Range")
	}
}

func finite(f float64) bool {
	return !math.IsNaN(f) && !math.IsInf(f, 0)
}

// snap panics if r isn't finite, has a negative size, or is out of int32 range.
func snap(r geom.Rect, inv float64) cellRange {
	minX, minY := r.Min()
	maxX, maxY := r.Max()

	if !finite(minX) || !finite(minY) || !finite(maxX) || !finite(maxY) {
		panic("hashgrid: rect contains non-finite values")
	}
	if maxX < minX || maxY < minY {
		panic("hashgrid: rect has a negative size")
	}

	x1, y1 := math.Floor(minX*inv), math.Floor(minY*inv)
	x2, y2 := math.Ceil(maxX*inv), math.Ceil(maxY*inv)

	if x2 == x1 {
		x2++
	}
	if y2 == y1 {
		y2++
	}

	if x1 < math.MinInt32 || y1 < math.MinInt32 || x2 > math.MaxInt32 || y2 > math.MaxInt32 {
		panic("hashgrid: rect is outside the grid's cell range")
	}

	return cellRange{minX: int32(x1), minY: int32(y1), maxX: int32(x2), maxY: int32(y2)}
}
