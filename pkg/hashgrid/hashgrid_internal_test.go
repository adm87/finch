package hashgrid

import (
	"math"
	"math/rand/v2"
	"slices"
	"testing"

	"github.com/adm87/finch/pkg/geom"
	"github.com/adm87/finch/pkg/slotmap"
)

type mover struct {
	rect   geom.Rect
	vx, vy float64
}

// The answer is worked out from scratch, so a bug shared with another grid can't hide.
func TestQueryMatchesBruteForce(t *testing.T) {
	const size = 64
	edges := []geom.Rect{
		{X: -64, Y: -64, Width: 64, Height: 64}, // exactly cell (-1, -1)
		{X: -64, Y: -64},                        // zero size on a cell corner
		{X: -0.5, Y: -0.5, Width: 1, Height: 1}, // straddles the origin
		{X: -1e-9, Width: 1e-9},                 // ends exactly on zero
		{X: -2048, Y: -2048, Width: 2048, Height: 2048},
		{X: -2048, Y: -2048, Width: 4096, Height: 4096},
	}

	rng := rand.New(rand.NewPCG(1, 2))
	ms := make([]mover, 3000)
	for i := range ms {
		s := 8 + rng.Float64()*40
		ms[i] = mover{
			rect: geom.Rect{X: -2048 + rng.Float64()*(4096-s), Y: -2048 + rng.Float64()*(4096-s), Width: s, Height: s},
			vx:   rng.Float64()*8 - 4,
			vy:   rng.Float64()*8 - 4,
		}
	}
	for i, r := range edges {
		ms[i] = mover{rect: r} // pinned in place
	}

	g := New[int](size, 0)
	keys := make([]Key, len(ms))
	for i, m := range ms {
		keys[i] = g.Insert(m.rect, i)
	}
	checkInvariants(t, g)

	for round := range 30 {
		for i := range ms {
			m := &ms[i]
			m.rect.X += m.vx
			m.rect.Y += m.vy
			if m.rect.X < -2048 || m.rect.X+m.rect.Width > 2048 {
				m.vx = -m.vx
			}
			if m.rect.Y < -2048 || m.rect.Y+m.rect.Height > 2048 {
				m.vy = -m.vy
			}
			g.Update(keys[i], m.rect)
		}
		for i := len(edges) + round*50; i < len(edges)+round*50+50; i++ {
			if got, ok := g.Delete(keys[i]); !ok || got != i {
				t.Fatalf("Delete = (%d, %v), want (%d, true)", got, ok, i)
			}
			keys[i] = g.Insert(ms[i].rect, i)
		}
		checkInvariants(t, g)
	}

	negative := 0
	for _, m := range ms {
		if m.rect.X < 0 && m.rect.Y < 0 {
			negative++
		}
	}
	if negative < len(ms)/5 {
		t.Fatalf("%d of %d items are fully negative, want at least a fifth", negative, len(ms))
	}

	qs := slices.Clone(edges)
	for range 300 {
		qs = append(qs, geom.Rect{X: -2048 + rng.Float64()*3900, Y: -2048 + rng.Float64()*3900, Width: 150, Height: 90})
	}
	for _, q := range qs {
		qc := cellsOf(q, size)
		var want []int
		for i, m := range ms {
			if overlaps(cellsOf(m.rect, size), qc) {
				want = append(want, i)
			}
		}

		got := g.Query(q, nil)
		slices.Sort(got)
		if !slices.Equal(got, want) {
			t.Errorf("Query(%+v) found %d items, want %d", q, len(got), len(want))
		}
	}

	for _, k := range keys {
		g.Delete(k)
	}
	if len(g.cells) != 0 || g.Len() != 0 {
		t.Errorf("after deleting everything, cells, Len() = %d, %d, want 0, 0", len(g.cells), g.Len())
	}
}

func TestRangeEndsEvenWhenStoppedEarly(t *testing.T) {
	g := New[int](10, 0)
	g.Insert(geom.Rect{Width: 5, Height: 5}, 1)

	for range g.Range(geom.Rect{Width: 5, Height: 5}) {
		for range g.Range(geom.Rect{}) {
			break
		}
		break
	}

	if g.ranging != 0 {
		t.Errorf("ranging = %d, want 0", g.ranging)
	}
}

// checkInvariants checks the bookkeeping that queries rely on without checking.
func checkInvariants[T any](t *testing.T, g *Grid[T]) {
	t.Helper()
	count := map[Key]int{}
	for c, entries := range g.cells {
		if len(entries) == 0 {
			t.Fatalf("cell %x is empty but still in the map", c)
		}
		x, y := int32(c>>32), int32(c)
		seen := map[Key]bool{}
		for _, e := range entries {
			if seen[e.key] {
				t.Fatalf("an item is in cell (%d, %d) twice", x, y)
			}
			seen[e.key] = true
			count[e.key]++

			it, ok := g.items.GetPtr(slotmap.Key(e.key))
			if !ok {
				t.Fatalf("cell (%d, %d) holds a deleted key", x, y)
			}
			if !it.cells.contains(x, y) {
				t.Fatalf("an item is in cell (%d, %d), outside its range", x, y)
			}
			if e.minX != it.cells.minX || e.minY != it.cells.minY {
				t.Fatalf("an entry in cell (%d, %d) has a stale min cell", x, y)
			}
		}
	}
	for k, it := range g.items.All() {
		want := (int64(it.cells.maxX) - int64(it.cells.minX)) * (int64(it.cells.maxY) - int64(it.cells.minY))
		if got := count[Key(k)]; int64(got) != want {
			t.Fatalf("an item is in %d cells, want %d", got, want)
		}
	}
}

// cellsOf works out the cells r covers the slow, obvious way.
func cellsOf(r geom.Rect, size float64) [4]float64 {
	c := [4]float64{
		math.Floor(r.X / size), math.Floor(r.Y / size),
		math.Ceil((r.X + r.Width) / size), math.Ceil((r.Y + r.Height) / size),
	}
	if c[2] == c[0] {
		c[2]++
	}
	if c[3] == c[1] {
		c[3]++
	}
	return c
}

func overlaps(a, b [4]float64) bool {
	return a[0] < b[2] && a[2] > b[0] && a[1] < b[3] && a[3] > b[1]
}
