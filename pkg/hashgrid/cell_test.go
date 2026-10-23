package hashgrid_test

import (
	"math"
	"math/rand/v2"
	"testing"

	"github.com/go-openapi/testify/require"

	"github.com/adm87/finch/pkg/hashgrid"
)

func newRange(b cells) hashgrid.CellRange {
	return hashgrid.NewCellRange(b.minX, b.minY, b.maxX, b.maxY)
}

func collect(cr hashgrid.CellRange) []hashgrid.Cell {
	var got []hashgrid.Cell
	for c := range cr.All() {
		got = append(got, c)
	}
	return got
}

func TestCell(t *testing.T) {
	// Packed values are literals so the test doesn't mirror the implementation.
	cases := []struct {
		name   string
		x, y   int32
		packed uint64
	}{
		{"origin", 0, 0, 0},
		{"x only", 1, 0, 1 << 32},
		{"y only", 0, 1, 1},
		{"negative one", -1, -1, 0xFFFFFFFF_FFFFFFFF},
		{"negative x", -1, 1, 0xFFFFFFFF_00000001},
		{"negative y", 1, -1, 0x00000001_FFFFFFFF},
		{"max", math.MaxInt32, math.MaxInt32, 0x7FFFFFFF_7FFFFFFF},
		{"min", math.MinInt32, math.MinInt32, 0x80000000_80000000},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			cell := hashgrid.PackCell(tc.x, tc.y)
			require.Equal(t, hashgrid.Cell(tc.packed), cell)

			x, y := hashgrid.UnpackCell(cell)
			require.Equal(t, [2]int32{tc.x, tc.y}, [2]int32{x, y})
		})
	}

	t.Run("round trip over full int32 range", func(t *testing.T) {
		rng := rand.New(rand.NewPCG(1, 2)) // fixed seed: reproducible
		for range 10_000 {
			x, y := int32(rng.Uint32()), int32(rng.Uint32())
			gx, gy := hashgrid.UnpackCell(hashgrid.PackCell(x, y))
			require.Equal(t, [2]int32{x, y}, [2]int32{gx, gy})
		}
	})
}

func TestCellRange_MinMax(t *testing.T) {
	cases := []struct {
		name string
		in   cells // canonical order
	}{
		{"positive", cells{1, 2, 3, 4}},
		{"negative", cells{-4, -3, -2, -1}},
		{"straddles zero", cells{-2, -1, 2, 1}},
		{"empty", cells{0, 0, 0, 0}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			forEachOrdering(t, tc.in, func(t *testing.T, b cells) {
				cr := newRange(b)
				minX, minY := cr.Min()
				maxX, maxY := cr.Max()
				require.Equal(t, tc.in, cells{minX, minY, maxX, maxY})
			})
		})
	}
}

func TestCellRange_Contains(t *testing.T) {
	r := cells{0, 0, 10, 10}
	neg := cells{-10, -10, 0, 0}

	cases := []struct {
		name string
		r    cells
		x, y int32
		want bool
	}{
		{"min corner is inclusive", r, 0, 0, true},
		{"last cell inside", r, 9, 9, true},
		{"center", r, 5, 5, true},
		{"max x is exclusive", r, 10, 5, false},
		{"max y is exclusive", r, 5, 10, false},
		{"left of min", r, -1, 5, false},
		{"below min", r, 5, -1, false},
		{"inside x, outside y", r, 5, 20, false},
		{"inside y, outside x", r, 20, 5, false},
		{"negative range, min corner", neg, -10, -10, true},
		{"negative range, max exclusive", neg, 0, 0, false},
		{"empty range contains nothing", cells{0, 0, 0, 0}, 0, 0, false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			require.Equal(t, tc.want, newRange(tc.r).Contains(tc.x, tc.y))
		})
	}
}

func TestCellRange_Overlaps(t *testing.T) {
	base := cells{0, 0, 10, 10}

	cases := []struct {
		name string
		a, b cells
		want bool
	}{
		{"partial overlap", base, cells{5, 5, 15, 15}, true},
		{"identical", base, base, true},
		{"b inside a", base, cells{2, 2, 4, 4}, true},
		{"cross shape", cells{0, 4, 10, 6}, cells{4, 0, 6, 10}, true},
		{"touch on x edge only", base, cells{10, 0, 20, 10}, false},
		{"touch on y edge only", base, cells{0, 10, 10, 20}, false},
		{"touch at corner only", base, cells{10, 10, 20, 20}, false},
		{"overlap in x, apart in y", base, cells{5, 20, 15, 30}, false},
		{"overlap in y, apart in x", base, cells{20, 5, 30, 15}, false},
		{"negative overlap", cells{-10, -10, 0, 0}, cells{-5, -5, 5, 5}, true},
		{"negative touch", cells{-10, -10, 0, 0}, cells{0, 0, 10, 10}, false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			a, b := newRange(tc.a), newRange(tc.b)
			require.Equal(t, tc.want, a.Overlaps(b))
			require.Equal(t, tc.want, b.Overlaps(a), "Overlaps must be symmetric")
		})
	}
}

func TestCellRange_All(t *testing.T) {
	P := hashgrid.PackCell

	cases := []struct {
		name string
		in   cells
		want []hashgrid.Cell
	}{
		{"empty", cells{0, 0, 0, 0}, nil},
		{"zero width", cells{0, 0, 0, 5}, nil},
		{"zero height", cells{0, 0, 5, 0}, nil},
		{"single cell", cells{3, 4, 4, 5}, []hashgrid.Cell{P(3, 4)}},
		{"2x2", cells{0, 0, 2, 2}, []hashgrid.Cell{P(0, 0), P(1, 0), P(0, 1), P(1, 1)}},
		{"3 wide, 2 tall", cells{0, 0, 3, 2}, []hashgrid.Cell{P(0, 0), P(1, 0), P(2, 0), P(0, 1), P(1, 1), P(2, 1)}},
		{"negative", cells{-1, -1, 1, 0}, []hashgrid.Cell{P(-1, -1), P(0, -1)}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			// Run every case in all four orderings: iteration must not depend on input order.
			forEachOrdering(t, tc.in, func(t *testing.T, b cells) {
				require.Equal(t, tc.want, collect(newRange(b)))
			})
		})
	}

	t.Run("early break stops iteration", func(t *testing.T) {
		var got []hashgrid.Cell
		for c := range newRange(cells{0, 0, 3, 3}).All() {
			got = append(got, c)
			if len(got) == 2 {
				break
			}
		}
		require.Equal(t, []hashgrid.Cell{P(0, 0), P(1, 0)}, got)
	})
}
