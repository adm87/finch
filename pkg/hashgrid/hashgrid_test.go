package hashgrid_test

import (
	"math"
	"slices"
	"testing"

	"github.com/adm87/finch/pkg/geom"
	"github.com/adm87/finch/pkg/hashgrid"
)

func TestNewPanicsOnBadSize(t *testing.T) {
	for _, size := range []float64{0, -1, math.NaN(), math.Inf(1)} {
		func() {
			defer func() {
				if recover() == nil {
					t.Errorf("New(%v) did not panic", size)
				}
			}()
			hashgrid.New[int](size, 0)
		}()
	}
}

func TestInsertPanicsOnBadRect(t *testing.T) {
	tests := []struct {
		name string
		r    geom.Rect
	}{
		{"NaN x", geom.Rect{X: math.NaN()}},
		{"Inf width", geom.Rect{Width: math.Inf(1)}},
		{"negative width", geom.Rect{Width: -1}},
		{"negative height", geom.Rect{Height: -1}},
		{"past int32", geom.Rect{X: 100 * math.MaxInt32}},
		{"before int32", geom.Rect{X: -100 * float64(math.MaxInt32)}},
		{"zero size on the last cell", geom.Rect{X: 10 * float64(math.MaxInt32)}}, // its max becomes MaxInt32+1
	}
	g := hashgrid.New[int](10, 0)
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defer func() {
				if recover() == nil {
					t.Errorf("Insert(%+v) did not panic", tt.r)
				}
			}()
			g.Insert(tt.r, 0)
		})
	}
}

func TestDeleteStalesKey(t *testing.T) {
	g := hashgrid.New[string](10, 0)
	k := g.Insert(geom.Rect{Width: 5, Height: 5}, "a")

	if got, ok := g.Delete(k); !ok || got != "a" {
		t.Fatalf("Delete = (%q, %v), want (a, true)", got, ok)
	}
	if _, ok := g.Delete(k); ok {
		t.Error("second Delete ok = true, want false")
	}
	if g.Update(k, geom.Rect{X: math.NaN()}) {
		t.Error("Update after Delete = true, want false")
	}

	k2 := g.Insert(geom.Rect{Width: 5, Height: 5}, "b")
	if g.Update(k, geom.Rect{X: 50}) {
		t.Error("stale Update after reuse = true, want false")
	}
	if got := g.Query(geom.Rect{Width: 5, Height: 5}, nil); !slices.Equal(got, []string{"b"}) {
		t.Errorf("Query = %v, want [b]", got)
	}

	g.Delete(k2)
	if g.Len() != 0 {
		t.Errorf("Len() = %d, want 0", g.Len())
	}
}

func TestMutationDuringRangePanics(t *testing.T) {
	g := hashgrid.New[int](10, 0)
	k := g.Insert(geom.Rect{Width: 5, Height: 5}, 1)
	all := geom.Rect{Width: 100, Height: 100}

	tests := []struct {
		name   string
		mutate func()
	}{
		{"Insert", func() { g.Insert(all, 2) }},
		{"Update", func() { g.Update(k, all) }},
		{"Delete", func() { g.Delete(k) }},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defer func() {
				if recover() == nil {
					t.Errorf("%s during Range did not panic", tt.name)
				}
			}()
			for range g.Range(all) {
				for range g.Range(all) {
					break
				}
				tt.mutate() // the outer Range is still running
			}
		})
	}

	if !g.Update(k, geom.Rect{X: 20}) {
		t.Error("Update after Range = false, want true")
	}
}

func TestRangeMatchesQuery(t *testing.T) {
	g := hashgrid.New[int](16, 0)
	for i := range 200 {
		x, y := float64(i*7%300)-150, float64(i*13%300)-150
		g.Insert(geom.Rect{X: x, Y: y, Width: float64(i % 40), Height: float64(i % 25)}, i)
	}
	q := geom.Rect{X: -60, Y: -40, Width: 130, Height: 90}

	var got []int
	for _, v := range g.Range(q) {
		got = append(got, v)
	}

	if want := g.Query(q, nil); !slices.Equal(got, want) {
		t.Errorf("Range = %v, want %v", got, want)
	}
}

func TestUpdateDoesNotAllocate(t *testing.T) {
	g := hashgrid.New[int](10, 100)
	keys := make([]hashgrid.Key, 100)
	for i := range keys {
		keys[i] = g.Insert(geom.Rect{X: float64(i * 3), Width: 4, Height: 4}, i)
	}

	var x float64
	move := func() {
		x += 7
		for i, k := range keys {
			g.Update(k, geom.Rect{X: math.Mod(float64(i*3)+x, 500), Width: 4, Height: 4})
		}
	}
	for range 200 {
		move() // fills the spare pool
	}

	if n := testing.AllocsPerRun(100, move); n != 0 {
		t.Errorf("Update allocates %v times per pass", n)
	}
}
