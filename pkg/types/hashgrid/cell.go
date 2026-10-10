package hashgrid

import (
	"errors"
	"iter"
)

var (
	ErrOversized = errors.New("oversized")
)

// Cell represents a single cell in the hashgrid.
// It is packed as a uint64 with the high 32 bits as the x coordinate
// and the low 32 bits as the y coordinate.
type Cell uint64

// PackCell packs the x and y coordinates into a Cell.
// The x coordinate is stored in the high 32 bits and the y coordinate in the low 32 bits.
func PackCell(x, y int32) Cell {
	return Cell(uint64(uint32(x))<<32 | uint64(uint32(y)))
}

// UnpackCell unpacks the x and y coordinates from a Cell.
// The x coordinate is retrieved from the high 32 bits and the y coordinate from the low 32 bits.
func UnpackCell(c Cell) (x, y int32) {
	x = int32(uint64(c) >> 32)
	y = int32(uint64(c) & 0xFFFFFFFF)
	return
}

// CellRange represents a rectangular range of cells in the hashgrid.
// It is defined by its minimum and maximum x and y coordinates.
type CellRange struct {
	minX, minY int32
	maxX, maxY int32
}

// NewCellRange creates a new CellRange with the given minimum and maximum x and y coordinates.
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

// Min returns the minimum x and y coordinates of the CellRange.
func (cr CellRange) Min() (x, y int32) {
	return cr.minX, cr.minY
}

// Max returns the maximum x and y coordinates of the CellRange.
func (cr CellRange) Max() (x, y int32) {
	return cr.maxX, cr.maxY
}

// Contains checks if the given x and y coordinates are within the CellRange.
// It returns true if the coordinates are inside the range, false otherwise.
func (cr CellRange) Contains(x, y int32) bool {
	return x >= cr.minX && x < cr.maxX && y >= cr.minY && y < cr.maxY
}

// Overlaps checks if the CellRange overlaps with another CellRange.
// It returns true if there is any intersection between the two ranges, false otherwise.
func (cr CellRange) Overlaps(other CellRange) bool {
	return cr.minX < other.maxX && cr.maxX > other.minX && cr.minY < other.maxY && cr.maxY > other.minY
}

// All returns an iterator over all cells within the CellRange.
// The iterator yields each cell as a Cell, starting from the minimum coordinates and
// iterating row by row until the maximum coordinates are reached.
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
