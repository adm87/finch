package hashgrid

import (
	"errors"
	"fmt"
	"iter"
	"math"
	"slices"

	"github.com/adm87/finch/pkg/slotmap"
)

var (
	ErrInsertFailed = errors.New("insert failed")
	ErrUpdateFailed = errors.New("update failed")
	ErrDeleteFailed = errors.New("delete failed")
)

type Region interface {
	Min() (float64, float64)
	Max() (float64, float64)
}

type number interface {
	~int | ~int32 | ~int64 | ~float32 | ~float64
}

func minmax[T number](a, b T) (T, T) {
	if a < b {
		return a, b
	}
	return b, a
}

type Padding struct {
	Left   int32
	Right  int32
	Top    int32
	Bottom int32
}

type item[T any] struct {
	value T
	cr    CellRange
}

type Key uint64

func (k Key) slot() slotmap.K {
	return slotmap.Unpack(uint64(k))
}

type rangeLock int

func (rl *rangeLock) Lock() {
	*rl += 1
}

func (rl *rangeLock) Unlock() {
	*rl -= 1
}

func (rl *rangeLock) Check() {
	if *rl != 0 {
		panic("hashgrid: mutation during Range (use Query to collect first)")
	}
}

type HashGrid[T any] struct {
	items    *slotmap.SlotMap[item[T]]
	grid     map[Cell][]Key
	padding  Padding
	cellSize int
	rl       rangeLock
}

func New[T any](cellSize int, initialCapacity int) *HashGrid[T] {
	if cellSize <= 0 {
		panic("cellSize must be greater than 0")
	}
	return &HashGrid[T]{
		items:    slotmap.New[item[T]](initialCapacity),
		grid:     make(map[Cell][]Key),
		cellSize: cellSize,
	}
}

func (hg *HashGrid[T]) WithPadding(padding Padding) *HashGrid[T] {
	if padding.Left < 0 || padding.Right < 0 || padding.Top < 0 || padding.Bottom < 0 {
		panic("padding values must be non-negative")
	}
	hg.padding = padding
	return hg
}

func (hg *HashGrid[T]) CellSize() int {
	return hg.cellSize
}

func (hg *HashGrid[T]) Padding() Padding {
	return hg.padding
}

func (hg *HashGrid[T]) Has(key Key) bool {
	_, err := hg.items.GetPtr(key.slot())
	return err == nil
}

func (hg *HashGrid[T]) GridLen() int {
	return len(hg.grid)
}

func (hg *HashGrid[T]) ItemCount() int {
	return hg.items.Len()
}

func (hg *HashGrid[T]) KeysAt(x, y int32) []Key {
	return slices.Clone(hg.grid[PackCell(x, y)])
}

func (hg *HashGrid[T]) Insert(value T, region Region) (Key, error) {
	hg.rl.Check()

	minX, minY := region.Min()
	maxX, maxY := region.Max()

	cx1, cy1, cx2, cy2 := hg.PaddedRegion(minX, minY, maxX, maxY)
	cr := NewCellRange(cx1, cy1, cx2, cy2)

	slot := hg.items.Insert(item[T]{
		value: value,
		cr:    cr,
	})

	key := Key(slot.Pack())
	hg.addToGrid(key, cr)

	return key, nil
}

func (hg *HashGrid[T]) Delete(key Key) error {
	hg.rl.Check()

	item, err := hg.items.Delete(key.slot())
	if err != nil {
		return fmt.Errorf("%w: %w", ErrDeleteFailed, err)
	}
	hg.removeFromGrid(key, item.cr)

	return nil
}

// Update moves key to region. It does nothing if the covered cells don't change.
func (hg *HashGrid[T]) Update(key Key, region Region) error {
	hg.rl.Check()

	item, err := hg.items.GetPtr(key.slot())
	if err != nil {
		return fmt.Errorf("%w: %w", ErrUpdateFailed, err)
	}

	minX, minY := region.Min()
	maxX, maxY := region.Max()

	cx1, cy1, cx2, cy2 := hg.PaddedRegion(minX, minY, maxX, maxY)
	cr := NewCellRange(cx1, cy1, cx2, cy2)

	if cr == item.cr {
		return nil // the item is still within the same cell range
	}

	old := item.cr
	for y := old.minY; y < old.maxY; y++ {
		for x := old.minX; x < old.maxX; x++ {
			if !cr.Contains(x, y) {
				hg.removeKey(PackCell(x, y), key)
			}
		}
	}
	for y := cr.minY; y < cr.maxY; y++ {
		for x := cr.minX; x < cr.maxX; x++ {
			if !old.Contains(x, y) {
				c := PackCell(x, y)
				hg.grid[c] = append(hg.grid[c], key)
			}
		}
	}
	item.cr = cr

	return nil
}

// Query appends to dst every entry whose padded cell range overlaps the cells
// covered by region. This is a broad phase: callers must test exact bounds.
// Pass dst[:0] to reuse a buffer. Safe to mutate the grid afterward.
func (hg *HashGrid[T]) Query(region Region, dst []T) []T {
	minX, minY := region.Min()
	maxX, maxY := region.Max()

	cx1, cy1, cx2, cy2 := hg.PaddedRegion(minX, minY, maxX, maxY)
	cr := NewCellRange(cx1, cy1, cx2, cy2)

	hg.each(cr, func(key Key, it *item[T]) bool {
		dst = append(dst, it.value)
		return true
	})

	return dst
}

// Range returns an iterator over all entries that intersect the given region.
// Insert, Update and Delete panic during iteration; use Query to collect first.
func (hg *HashGrid[T]) Range(region Region) iter.Seq2[Key, T] {
	minX, minY := region.Min()
	maxX, maxY := region.Max()

	cx1, cy1, cx2, cy2 := hg.PaddedRegion(minX, minY, maxX, maxY)
	cr := NewCellRange(cx1, cy1, cx2, cy2)

	return func(yield func(Key, T) bool) {
		hg.rl.Lock()
		defer hg.rl.Unlock()

		hg.each(cr, func(key Key, it *item[T]) bool {
			return yield(key, it.value)
		})
	}
}

func (hg *HashGrid[T]) PaddedRegion(minX, minY, maxX, maxY float64) (int32, int32, int32, int32) {
	cx1, cy1, cx2, cy2 := hg.Snap(minX, minY, maxX, maxY)
	return cx1 - hg.padding.Left,
		cy1 - hg.padding.Top,
		cx2 + hg.padding.Right,
		cy2 + hg.padding.Bottom
}

func (hg *HashGrid[T]) Snap(minX, minY, maxX, maxY float64) (int32, int32, int32, int32) {
	cellSize := float64(hg.cellSize)

	minCellX, maxCellX := minmax(minX, maxX)
	minCellY, maxCellY := minmax(minY, maxY)

	x1 := int32(math.Floor(minCellX / cellSize))
	y1 := int32(math.Floor(minCellY / cellSize))
	x2 := int32(math.Ceil(maxCellX / cellSize))
	y2 := int32(math.Ceil(maxCellY / cellSize))

	// A zero-size region still covers one cell.
	if x1 == x2 {
		x2++
	}
	if y1 == y2 {
		y2++
	}

	return x1, y1, x2, y2
}

func (hg *HashGrid[T]) addToGrid(key Key, cr CellRange) {
	for cell := range cr.All() {
		hg.grid[cell] = append(hg.grid[cell], key)
	}
}

func (hg *HashGrid[T]) removeFromGrid(key Key, cr CellRange) {
	for cell := range cr.All() {
		hg.removeKey(cell, key)
	}
}

func (hg *HashGrid[T]) removeKey(cell Cell, key Key) {
	keys, ok := hg.grid[cell]
	if !ok {
		return
	}

	for i, k := range keys {
		if k == key {
			l := len(keys) - 1
			keys[i] = keys[l]
			keys = keys[:l]
			break
		}
	}

	if len(keys) == 0 {
		// Deleting empty cells keeps GridLen accurate but churns cells that refill often.
		delete(hg.grid, cell)
		return
	}

	hg.grid[cell] = keys
}

func (hg *HashGrid[T]) each(cr CellRange, fn func(key Key, it *item[T]) bool) {
	for y := cr.minY; y < cr.maxY; y++ {
		for x := cr.minX; x < cr.maxX; x++ {
			for _, key := range hg.grid[PackCell(x, y)] {
				it, err := hg.items.GetPtr(key.slot())
				if err != nil {
					continue
				}

				refX := max(it.cr.minX, cr.minX)
				refY := max(it.cr.minY, cr.minY)

				if x != refX || y != refY {
					continue
				}

				if !fn(key, it) {
					return
				}
			}
		}
	}
}
