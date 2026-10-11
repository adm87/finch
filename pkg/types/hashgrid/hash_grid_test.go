package hashgrid_test

import (
	"testing"

	"github.com/adm87/finch/pkg/types/hashgrid"
	"github.com/go-openapi/testify/require"
)

type bounds[T ~int32 | ~float64] struct{ minX, minY, maxX, maxY T }

func (b bounds[T]) Min() (T, T) {
	return b.minX, b.minY
}

func (b bounds[T]) Max() (T, T) {
	return b.maxX, b.maxY
}

type (
	region = bounds[float64] // world-space input for Snap / PaddedRegion
	cells  = bounds[int32]   // cell-space input for CellRange
)

// forEachOrdering runs fn for b as given, then with x inverted, y inverted,
// and both inverted. Every function that takes a min/max pair should return
// the same result for all four.
func forEachOrdering[T ~int32 | ~float64](t *testing.T, b bounds[T], fn func(*testing.T, bounds[T])) {
	t.Helper()
	orderings := []struct {
		name string
		b    bounds[T]
	}{
		{"ordered", b},
		{"x inverted", bounds[T]{b.maxX, b.minY, b.minX, b.maxY}},
		{"y inverted", bounds[T]{b.minX, b.maxY, b.maxX, b.minY}},
		{"both inverted", bounds[T]{b.maxX, b.maxY, b.minX, b.minY}},
	}
	for _, o := range orderings {
		t.Run(o.name, func(t *testing.T) { fn(t, o.b) })
	}
}

func TestHashGrid_Initialization(t *testing.T) {
	t.Run("basic initialization", func(t *testing.T) {
		hg := hashgrid.New[int](10, 100)
		require.NotNil(t, hg)
		require.Equal(t, 10, hg.CellSize())
	})

	t.Run("initialization with padding", func(t *testing.T) {
		hg := hashgrid.New[int](10, 100).WithPadding(hashgrid.Padding{Left: 1, Right: 1, Top: 1, Bottom: 1})
		require.Equal(t, hashgrid.Padding{Left: 1, Right: 1, Top: 1, Bottom: 1}, hg.Padding())
	})

	t.Run("panic on negative padding", func(t *testing.T) {
		require.Panics(t, func() {
			hashgrid.New[int](10, 100).WithPadding(hashgrid.Padding{Left: -1})
		})
		require.Panics(t, func() {
			hashgrid.New[int](10, 100).WithPadding(hashgrid.Padding{Right: -1})
		})
		require.Panics(t, func() {
			hashgrid.New[int](10, 100).WithPadding(hashgrid.Padding{Top: -1})
		})
		require.Panics(t, func() {
			hashgrid.New[int](10, 100).WithPadding(hashgrid.Padding{Bottom: -1})
		})
	})

	t.Run("panic on invalid cell size", func(t *testing.T) {
		require.Panics(t, func() {
			hashgrid.New[int](0, 100)
		})
	})
}

func TestHashGrid_Snap(t *testing.T) {
	hg := hashgrid.New[int](10, 100)

	cases := []struct {
		name string
		in   region // canonical order: min <= max
		want [4]int32
	}{
		// Rounding: floor on min, ceil on max
		{"positive fractional", region{5, 5, 25, 25}, [4]int32{0, 0, 3, 3}},
		{"non-uniform fractional", region{3, 7, 22, 18}, [4]int32{0, 0, 3, 2}},
		{"negative fractional", region{-25, -25, -5, -5}, [4]int32{-3, -3, 0, 0}},
		{"straddles zero", region{-15, -5, 15, 5}, [4]int32{-2, -1, 2, 1}},
		{"mixed sign per axis", region{5, -15, 15, -5}, [4]int32{0, -2, 2, 0}},

		// Exact multiples of the cell size
		{"positive on boundaries", region{0, 0, 20, 20}, [4]int32{0, 0, 2, 2}},
		{"negative on boundaries", region{-20, -20, 0, 0}, [4]int32{-2, -2, 0, 0}},
		{"max on boundary does not spill", region{0, 0, 10, 10}, [4]int32{0, 0, 1, 1}},
		{"min on boundary", region{10, 10, 20, 20}, [4]int32{1, 1, 2, 2}},

		// Zero-size inside a cell: floor != ceil, so no expansion is needed
		{"zero-size inside cell", region{5, 5, 5, 5}, [4]int32{0, 0, 1, 1}},
		{"zero-size negative inside cell", region{-5, -5, -5, -5}, [4]int32{-1, -1, 0, 0}},

		// Zero-size on a boundary: floor == ceil, so the expansion branch runs
		{"zero-size at origin", region{0, 0, 0, 0}, [4]int32{0, 0, 1, 1}},
		{"zero-size on boundary", region{10, 10, 10, 10}, [4]int32{1, 1, 2, 2}},
		{"zero-size negative on boundary", region{-10, -10, -10, -10}, [4]int32{-1, -1, 0, 0}},

		// Each axis expands independently
		{"zero width on boundary", region{10, 5, 10, 25}, [4]int32{1, 0, 2, 3}},
		{"zero height on boundary", region{5, 10, 25, 10}, [4]int32{0, 1, 3, 2}},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			forEachOrdering(t, tc.in, func(t *testing.T, r region) {
				minX, minY, maxX, maxY := hg.Snap(r.minX, r.minY, r.maxX, r.maxY)
				require.Equal(t, tc.want, [4]int32{minX, minY, maxX, maxY})
			})
		})
	}
}

func TestHashGrid_PaddedRegion(t *testing.T) {
	distinct := hashgrid.Padding{Left: 1, Right: 2, Top: 3, Bottom: 4}
	uniform := hashgrid.Padding{Left: 1, Right: 1, Top: 1, Bottom: 1}

	cases := []struct {
		name string
		pad  hashgrid.Padding
		in   region // canonical order: min <= max
		want [4]int32
	}{
		// Distinct value per side, so any swapped edge fails
		{"each side on correct edge", distinct, region{5, 5, 25, 25}, [4]int32{-1, -3, 5, 7}},
		{"negative coordinates", distinct, region{-25, -25, -5, -5}, [4]int32{-4, -6, 2, 4}},

		// Zero padding is a no-op over Snap
		{"zero padding", hashgrid.Padding{}, region{5, 5, 25, 25}, [4]int32{0, 0, 3, 3}},

		// Padding is applied after Snap's zero-size expansion
		{"zero-size inside cell", uniform, region{5, 5, 5, 5}, [4]int32{-1, -1, 2, 2}},
		{"zero-size on boundary", uniform, region{10, 10, 10, 10}, [4]int32{0, 0, 3, 3}},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			hg := hashgrid.New[int](10, 100).WithPadding(tc.pad)
			forEachOrdering(t, tc.in, func(t *testing.T, r region) {
				minX, minY, maxX, maxY := hg.PaddedRegion(r.minX, r.minY, r.maxX, r.maxY)
				require.Equal(t, tc.want, [4]int32{minX, minY, maxX, maxY})
			})
		})
	}
}

func TestHashGrid_Insert(t *testing.T) {
	t.Run("insert single item", func(t *testing.T) {
		hg := hashgrid.New[int](10, 100)
		key, err := hg.Insert(42, bounds[float64]{minX: 0, minY: 0, maxX: 10, maxY: 10})
		require.NoError(t, err)
		require.True(t, hg.Has(key))
		require.Equal(t, hg.GridLen(), 1)
		require.Equal(t, hg.ItemCount(), 1)
	})

	t.Run("insert multiple items at different cells", func(t *testing.T) {
		hg := hashgrid.New[int](10, 100)
		key1, err := hg.Insert(42, bounds[float64]{minX: 0, minY: 0, maxX: 10, maxY: 10})
		require.NoError(t, err)
		require.True(t, hg.Has(key1))

		key2, err := hg.Insert(84, bounds[float64]{minX: 10, minY: 10, maxX: 20, maxY: 20})
		require.NoError(t, err)
		require.True(t, hg.Has(key2))

		require.Equal(t, hg.GridLen(), 2)
		require.Equal(t, hg.ItemCount(), 2)
	})

	t.Run("insert multiple items at the same cell", func(t *testing.T) {
		hg := hashgrid.New[int](10, 100)
		key1, err := hg.Insert(42, bounds[float64]{minX: 0, minY: 0, maxX: 10, maxY: 10})
		require.NoError(t, err)
		require.True(t, hg.Has(key1))

		key2, err := hg.Insert(84, bounds[float64]{minX: 0, minY: 0, maxX: 10, maxY: 10})
		require.NoError(t, err)
		require.True(t, hg.Has(key2))

		require.Equal(t, hg.GridLen(), 1)
		require.Equal(t, hg.ItemCount(), 2)
	})

	t.Run("insert multiple items with overlapping cells", func(t *testing.T) {
		hg := hashgrid.New[int](10, 100)
		key1, err := hg.Insert(42, bounds[float64]{minX: 0, minY: 0, maxX: 15, maxY: 15})
		require.NoError(t, err)
		require.True(t, hg.Has(key1))

		key2, err := hg.Insert(84, bounds[float64]{minX: 10, minY: 10, maxX: 20, maxY: 20})
		require.NoError(t, err)
		require.True(t, hg.Has(key2))

		require.Equal(t, hg.GridLen(), 4)
		require.Equal(t, hg.ItemCount(), 2)
	})

	t.Run("insert item with zero-size region", func(t *testing.T) {
		hg := hashgrid.New[int](10, 100)
		key, err := hg.Insert(42, region{minX: 5, minY: 5, maxX: 5, maxY: 5})
		require.NoError(t, err)
		require.True(t, hg.Has(key))
		require.Equal(t, hg.GridLen(), 1)
		require.Equal(t, hg.ItemCount(), 1)
	})

	t.Run("insert item with negative coordinates at a single cell", func(t *testing.T) {
		hg := hashgrid.New[int](10, 100)
		key, err := hg.Insert(42, region{minX: -5, minY: -5, maxX: 0, maxY: 0})
		require.NoError(t, err)
		require.True(t, hg.Has(key))
		require.Equal(t, hg.GridLen(), 1)
		require.Equal(t, hg.ItemCount(), 1)
	})

	t.Run("insert item with negative coordinates at multiple cells", func(t *testing.T) {
		hg := hashgrid.New[int](10, 100)
		key, err := hg.Insert(42, region{minX: -5, minY: -5, maxX: 5, maxY: 5})
		require.NoError(t, err)
		require.True(t, hg.Has(key))
		require.Equal(t, hg.GridLen(), 4)
		require.Equal(t, hg.ItemCount(), 1)
	})
}

func TestHashGrid_Delete(t *testing.T) {
	t.Run("delete existing item", func(t *testing.T) {
		hg := hashgrid.New[int](10, 100)
		key, err := hg.Insert(42, region{minX: 0, minY: 0, maxX: 10, maxY: 10})
		require.NoError(t, err)
		require.True(t, hg.Has(key))

		err = hg.Delete(key)
		require.NoError(t, err)
		require.False(t, hg.Has(key))
		require.Equal(t, hg.GridLen(), 0)
		require.Equal(t, hg.ItemCount(), 0)
	})

	t.Run("delete non-existing item", func(t *testing.T) {
		hg := hashgrid.New[int](10, 100)
		err := hg.Delete(999)
		require.Error(t, err)
	})

	t.Run("delete item with shared cells", func(t *testing.T) {
		hg := hashgrid.New[int](10, 100)
		key1, err := hg.Insert(42, region{minX: 0, minY: 0, maxX: 10, maxY: 10})
		require.NoError(t, err)
		require.True(t, hg.Has(key1))

		key2, err := hg.Insert(84, region{minX: 0, minY: 0, maxX: 10, maxY: 10})
		require.NoError(t, err)
		require.True(t, hg.Has(key2))

		err = hg.Delete(key1)
		require.NoError(t, err)
		require.False(t, hg.Has(key1))
		require.True(t, hg.Has(key2))
		require.Equal(t, hg.GridLen(), 1)
		require.Equal(t, hg.ItemCount(), 1)
	})

	t.Run("delete item spanning multiple shared cells", func(t *testing.T) {
		hg := hashgrid.New[int](10, 100)
		key1, err := hg.Insert(42, region{minX: 0, minY: 0, maxX: 20, maxY: 20})
		require.NoError(t, err)
		require.True(t, hg.Has(key1))

		key2, err := hg.Insert(84, region{minX: 10, minY: 10, maxX: 20, maxY: 20})
		require.NoError(t, err)
		require.True(t, hg.Has(key2))

		err = hg.Delete(key1)
		require.NoError(t, err)
		require.False(t, hg.Has(key1))
		require.True(t, hg.Has(key2))
		require.Equal(t, hg.GridLen(), 1)
		require.Equal(t, hg.ItemCount(), 1)
	})
}

func TestHashGrid_Update(t *testing.T) {
	t.Run("update existing item", func(t *testing.T) {
		hg := hashgrid.New[int](10, 100)
		key, err := hg.Insert(42, region{minX: 0, minY: 0, maxX: 10, maxY: 10})
		require.NoError(t, err)
		require.True(t, hg.Has(key))

		err = hg.Update(key, region{minX: 10, minY: 10, maxX: 20, maxY: 20})
		require.NoError(t, err)
		require.True(t, hg.Has(key))
		require.Equal(t, hg.GridLen(), 1)
		require.Equal(t, hg.ItemCount(), 1)
	})

	t.Run("update non-existing item", func(t *testing.T) {
		hg := hashgrid.New[int](10, 100)
		err := hg.Update(999, region{minX: 0, minY: 0, maxX: 10, maxY: 10})
		require.Error(t, err)
	})

	t.Run("update item spanning multiple shared cells", func(t *testing.T) {
		hg := hashgrid.New[int](10, 100)
		key, err := hg.Insert(42, region{minX: 0, minY: 0, maxX: 20, maxY: 20})
		require.NoError(t, err)
		require.True(t, hg.Has(key))

		err = hg.Update(key, region{minX: 10, minY: 10, maxX: 20, maxY: 20})
		require.NoError(t, err)
		require.True(t, hg.Has(key))
		require.Equal(t, hg.GridLen(), 1)
		require.Equal(t, hg.ItemCount(), 1)
	})
}
