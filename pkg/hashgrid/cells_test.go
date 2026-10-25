package hashgrid_test

import (
	"testing"

	"github.com/go-openapi/testify/require"

	"github.com/adm87/finch/pkg/geom"
	"github.com/adm87/finch/pkg/hashgrid"
)

func rect(minX, minY, maxX, maxY float64) geom.Rect {
	return geom.Rect{X: minX, Y: minY, Width: maxX - minX, Height: maxY - minY}
}

func collectCells(hg *hashgrid.HashGrid[int], region geom.Rect) map[[2]int32]int {
	out := make(map[[2]int32]int)
	for cell, n := range hg.Cells(region) {
		x, y := hashgrid.UnpackCell(cell)
		out[[2]int32{x, y}] = n
	}
	return out
}

func newCellsGrid(t *testing.T) *hashgrid.HashGrid[int] {
	t.Helper()
	hg := hashgrid.New[int](10, 8)
	_, err := hg.Insert(1, rect(5, 5, 25, 15))
	require.NoError(t, err)
	_, err = hg.Insert(2, rect(0, 0, 5, 5))
	require.NoError(t, err)
	_, err = hg.Insert(3, rect(100, 100, 105, 105))
	require.NoError(t, err)
	return hg
}

func TestCellsInSmallRegion(t *testing.T) {
	hg := newCellsGrid(t)

	require.Equal(t, map[[2]int32]int{
		{0, 0}: 2, {1, 0}: 1, {2, 0}: 1,
		{0, 1}: 1, {1, 1}: 1, {2, 1}: 1,
	}, collectCells(hg, rect(0, 0, 30, 20)))

	require.Empty(t, collectCells(hg, rect(50, 50, 60, 60)))
}

func TestCellsInRegionLargerThanGrid(t *testing.T) {
	hg := newCellsGrid(t)

	got := collectCells(hg, rect(-1e6, -1e6, 1e6, 1e6))
	require.Len(t, got, 7)
	require.Equal(t, 1, got[[2]int32{10, 10}])
	require.Equal(t, 2, got[[2]int32{0, 0}])
}

func TestCellsStopsEarlyAndLocksTheGrid(t *testing.T) {
	hg := newCellsGrid(t)

	visited := 0
	for range hg.Cells(rect(-1e6, -1e6, 1e6, 1e6)) {
		visited++
		require.Panics(t, func() { _, _ = hg.Insert(9, rect(0, 0, 1, 1)) })
		break
	}
	require.Equal(t, 1, visited)

	_, err := hg.Insert(9, rect(0, 0, 1, 1))
	require.NoError(t, err)
}

func TestCellBounds(t *testing.T) {
	hg := hashgrid.New[int](10, 1)
	minX, minY, maxX, maxY := hg.CellBounds(hashgrid.PackCell(-1, 2))
	require.Equal(t, []float64{-10, 20, 0, 30}, []float64{minX, minY, maxX, maxY})
}
