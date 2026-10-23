package hashgrid

import (
	"errors"
	"iter"
)

var (
	ErrOversized = errors.New("oversized")
)

// Cell is a grid coordinate packed with x in the high 32 bits and y in the low 32 bits.
type Cell uint64

func PackCell(x, y int32) Cell {
	return Cell(uint64(uint32(x))<<32 | uint64(uint32(y)))
}

func UnpackCell(c Cell) (x, y int32) {
	x = int32(uint64(c) >> 32)
	y = int32(uint64(c) & 0xFFFFFFFF)
	return
}

// CellRange is a rectangle of cells. It includes its min edge and excludes its max edge.
type CellRange struct {
	minX, minY int32
	maxX, maxY int32
}

// NewCellRange accepts the corners in either order.
func NewCellRange(minX, minY, maxX, maxY int32) CellRange {
	x1, x2 := minmax(minX, maxX)
	y1, y2 := minmax(minY, maxY)
	return CellRange{
		minX: x1,
		minY: y1,
		maxX: x2,
		maxY: y2,
	}
}

func (cr CellRange) Min() (x, y int32) {
	return cr.minX, cr.minY
}

func (cr CellRange) Max() (x, y int32) {
	return cr.maxX, cr.maxY
}

func (cr CellRange) Contains(x, y int32) bool {
	return x >= cr.minX && x < cr.maxX && y >= cr.minY && y < cr.maxY
}

func (cr CellRange) Overlaps(other CellRange) bool {
	return cr.minX < other.maxX && cr.maxX > other.minX && cr.minY < other.maxY && cr.maxY > other.minY
}

// All iterates over the cells row by row.
func (cr CellRange) All() iter.Seq[Cell] {
	return func(yield func(Cell) bool) {
		for y := cr.minY; y < cr.maxY; y++ {
			for x := cr.minX; x < cr.maxX; x++ {
				if !yield(PackCell(x, y)) {
					return
				}
			}
		}
	}
}
