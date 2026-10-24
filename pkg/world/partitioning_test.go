package world

import (
	"math"
	"math/rand/v2"
	"slices"
	"testing"

	"github.com/yohamta/donburi"

	"github.com/adm87/finch/pkg/geom"
	"github.com/adm87/finch/pkg/hashgrid"
	"github.com/adm87/finch/pkg/transform"
)

// Helpers are prefixed with "pt" so they can't collide with other test files.

var ptResolutions = []int{16, 32, 64}

// ptNewEntry creates an unindexed entity with the given local rect.
func ptNewEntry(t *testing.T, w donburi.World, rect geom.Rect, extra ...donburi.IComponentType) *donburi.Entry {
	t.Helper()

	entry := w.Entry(w.Create(aggregateComponentTypes(extra...)...))
	if t, ok := transform.GetTransform(entry); ok {
		t.SetScale(1, 1)
	}

	b, bOk := transform.GetBounds(entry)
	if !bOk {
		t.Fatalf("expected entry to have bounds component")
	}

	b.SetRect(rect)
	return entry
}

func ptMove(entry *donburi.Entry, x, y float64) {
	t, _ := transform.GetTransform(entry)
	t.SetPosition(x, y)
}

func ptRect(x, y, w, h float64) geom.Rect {
	return geom.Rect{X: x, Y: y, Width: w, Height: h}
}

// ptHits collects entities from every grid for region.
func ptHits(p *partitions, region geom.Rect) []donburi.Entity {
	var out []donburi.Entity
	for _, g := range p.grids {
		out = g.Query(region, out)
	}
	return out
}

func ptHitSet(p *partitions, region geom.Rect) map[donburi.Entity]struct{} {
	out := make(map[donburi.Entity]struct{}, len(ptHits(p, region)))
	for _, e := range ptHits(p, region) {
		out[e] = struct{}{}
	}
	return out
}

func containsEntity(set map[donburi.Entity]struct{}, v donburi.Entity) bool {
	_, ok := set[v]
	return ok
}

func ptCounts(p *partitions) []int {
	out := make([]int, len(p.grids))
	for i, g := range p.grids {
		out[i] = g.ItemCount()
	}
	return out
}

func ptIndex(t *testing.T, entry *donburi.Entry) *indexing {
	t.Helper()
	idx, ok := getIndexing(entry)
	if !ok {
		t.Fatal("entry has no indexing component")
	}
	return idx
}

func ptMustAdd(t *testing.T, p *partitions, entry *donburi.Entry) {
	t.Helper()
	if err := p.add(entry); err != nil {
		t.Fatalf("add failed: %v", err)
	}
}

func ptMustUpdate(t *testing.T, p *partitions, entry *donburi.Entry) {
	t.Helper()
	if err := p.update(entry); err != nil {
		t.Fatalf("update failed: %v", err)
	}
}

// ptCellKeyTotal counts the keys stored in g's cells within [lo, hi) on both axes.
// It catches keys left behind in old cells.
func ptCellKeyTotal(g *hashgrid.HashGrid[donburi.Entity], lo, hi float64) int {
	cs := float64(g.CellSize())
	from, to := int32(math.Floor(lo/cs)), int32(math.Ceil(hi/cs))

	total := 0
	for y := from; y < to; y++ {
		for x := from; x < to; x++ {
			total += len(g.KeysAt(x, y))
		}
	}
	return total
}

// ---------------------------------------------------------------------------
// Construction
// ---------------------------------------------------------------------------

func TestNewPartitions(t *testing.T) {
	t.Run("panics on empty resolutions", func(t *testing.T) {
		defer func() {
			if recover() == nil {
				t.Fatal("expected panic")
			}
		}()
		newPartitions(nil)
	})

	t.Run("panics on zero resolution", func(t *testing.T) {
		defer func() {
			if recover() == nil {
				t.Fatal("expected panic")
			}
		}()
		newPartitions([]int{16, 0, 32})
	})

	t.Run("panics on negative resolution", func(t *testing.T) {
		defer func() {
			if recover() == nil {
				t.Fatal("expected panic")
			}
		}()
		newPartitions([]int{-1})
	})

	t.Run("sorts ascending without mutating the input", func(t *testing.T) {
		in := []int{64, 16, 32}
		p := newPartitions(in)

		got := make([]int, len(p.grids))
		for i, g := range p.grids {
			got[i] = g.CellSize()
		}
		if !slices.Equal(got, []int{16, 32, 64}) {
			t.Fatalf("expected ascending cell sizes, got %v", got)
		}
		if !slices.Equal(in, []int{64, 16, 32}) {
			t.Fatalf("input slice was modified: %v", in)
		}
	})
}

func TestNearestFit(t *testing.T) {
	p := newPartitions(ptResolutions)

	cases := []struct {
		name string
		w, h float64
		want int
	}{
		{"zero size", 0, 0, 0},
		{"fits finest", 12, 8, 0},
		{"exact finest boundary", 16, 4, 0},
		{"just over finest", 16.01, 4, 1},
		{"uses the larger side", 4, 40, 2},
		{"exact coarsest boundary", 64, 64, 2},
		{"larger than coarsest", 500, 10, 2},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := p.nearestFit(geom.Rect{X: 0, Y: 0, Width: tc.w, Height: tc.h}); got != tc.want {
				t.Fatalf("nearestFit(%vx%v) = %d, want %d", tc.w, tc.h, got, tc.want)
			}
		})
	}
}

func TestNearestFitNonFinite(t *testing.T) {
	p := newPartitions(ptResolutions)
	last := len(p.grids) - 1

	for _, v := range []float64{math.NaN(), math.Inf(1)} {
		if got := p.nearestFit(geom.Rect{X: 0, Y: 0, Width: v, Height: 1}); got != last {
			t.Fatalf("nearestFit with width %v = %d, want %d", v, got, last)
		}
	}
}

// ---------------------------------------------------------------------------
// Lifecycle
// ---------------------------------------------------------------------------

func TestFreshEntryIsUnindexed(t *testing.T) {
	w := donburi.NewWorld()
	entry := ptNewEntry(t, w, geom.Rect{X: 0, Y: 0, Width: 12, Height: 10})

	idx := ptIndex(t, entry)
	if idx.partition != -1 || idx.key != 0 {
		t.Fatalf("expected a fresh entry to be unindexed (-1, 0), got partition=%d key=%v", idx.partition, idx.key)
	}
}

func TestAddIndexesEntry(t *testing.T) {
	w := donburi.NewWorld()
	p := newPartitions(ptResolutions)
	entry := ptNewEntry(t, w, geom.Rect{X: 0, Y: 0, Width: 12, Height: 10})

	ptMustAdd(t, p, entry)

	idx := ptIndex(t, entry)
	if idx.partition != 0 || idx.key == 0 {
		t.Fatalf("expected partition 0 with a non-zero key, got partition=%d key=%v", idx.partition, idx.key)
	}
	if got := ptCounts(p); !slices.Equal(got, []int{1, 0, 0}) {
		t.Fatalf("unexpected item counts: %v", got)
	}

	hits := ptHits(p, ptRect(0, 0, 16, 16))
	if len(hits) != 1 || hits[0] != entry.Entity() {
		t.Fatalf("expected exactly the entry, got %v", hits)
	}
}

func TestAddTwiceDoesNotDuplicate(t *testing.T) {
	w := donburi.NewWorld()
	p := newPartitions(ptResolutions)
	entry := ptNewEntry(t, w, ptRect(0, 0, 12, 10))

	ptMustAdd(t, p, entry)
	ptMustAdd(t, p, entry)

	if got := ptCounts(p); !slices.Equal(got, []int{1, 0, 0}) {
		t.Fatalf("expected a single item, got %v", got)
	}
	if hits := ptHits(p, ptRect(0, 0, 16, 16)); len(hits) != 1 {
		t.Fatalf("expected the entry once, found %d", len(hits))
	}
}

func TestUpdateSamePartitionKeepsKey(t *testing.T) {
	w := donburi.NewWorld()
	p := newPartitions(ptResolutions)
	entry := ptNewEntry(t, w, ptRect(0, 0, 12, 10))
	ptMustAdd(t, p, entry)
	before := *ptIndex(t, entry)

	ptMove(entry, 5, 5)
	ptMustUpdate(t, p, entry)

	after := ptIndex(t, entry)
	if after.partition != before.partition || after.key != before.key {
		t.Fatalf("expected the same partition and key, before=%+v after=%+v", before, *after)
	}
	if hits := ptHits(p, ptRect(5, 5, 12, 10)); len(hits) != 1 {
		t.Fatalf("expected the entry at its new position, found %d", len(hits))
	}
}

func TestMoveLeavesNoResidue(t *testing.T) {
	w := donburi.NewWorld()
	p := newPartitions(ptResolutions)
	entry := ptNewEntry(t, w, ptRect(0, 0, 12, 10))
	ptMustAdd(t, p, entry)

	origin := ptRect(0, 0, 16, 16)
	far := ptRect(1000, 1000, 16, 16)

	if len(ptHits(p, origin)) != 1 {
		t.Fatal("expected the entry at its start")
	}

	ptMove(entry, 1000, 1000)
	ptMustUpdate(t, p, entry)

	if got := ptHits(p, origin); len(got) != 0 {
		t.Fatalf("entry still found at its old location: %v", got)
	}
	if got := ptHits(p, far); len(got) != 1 {
		t.Fatalf("expected the entry at its new location, found %d", len(got))
	}
}

func TestResizeAcrossPartitions(t *testing.T) {
	w := donburi.NewWorld()
	p := newPartitions(ptResolutions)
	entry := ptNewEntry(t, w, ptRect(0, 0, 12, 10))
	ptMustAdd(t, p, entry)

	steps := []struct {
		rect geom.Rect
		want int
	}{
		{ptRect(0, 0, 48, 40), 2}, // grow
		{ptRect(0, 0, 24, 20), 1}, // shrink one step
		{ptRect(0, 0, 12, 10), 0}, // back to the start
		{ptRect(0, 0, 90, 70), 2}, // larger than every cell size
	}

	region := ptRect(0, 0, 256, 256)
	for _, s := range steps {
		b, bOk := transform.GetBounds(entry)
		if !bOk {
			t.Fatalf("expected entry to have bounds component")
		}

		b.SetRect(s.rect)
		ptMustUpdate(t, p, entry)

		idx := ptIndex(t, entry)
		if idx.partition != s.want || idx.key == 0 {
			t.Fatalf("%+v: expected partition %d with a key, got partition=%d key=%v", s.rect, s.want, idx.partition, idx.key)
		}

		counts := ptCounts(p)
		for i, c := range counts {
			want := 0
			if i == s.want {
				want = 1
			}
			if c != want {
				t.Fatalf("%+v: expected one item in partition %d, got counts %v", s.rect, s.want, counts)
			}
		}
		if hits := ptHits(p, region); len(hits) != 1 {
			t.Fatalf("%+v: expected the entry exactly once, found %d", s.rect, len(hits))
		}
	}
}

func TestMultiCellItemIsReturnedOnce(t *testing.T) {
	w := donburi.NewWorld()
	p := newPartitions(ptResolutions)

	// 40x40 at (50,50) lands in the 64 grid and straddles the x=64 and y=64 cell edges.
	entry := ptNewEntry(t, w, ptRect(50, 50, 40, 40))
	ptMustAdd(t, p, entry)

	if got := ptIndex(t, entry).partition; got != 2 {
		t.Fatalf("expected partition 2, got %d", got)
	}

	regions := map[string]struct {
		rect geom.Rect
		want int
	}{
		"covers every cell":  {ptRect(0, 0, 256, 256), 1},
		"touches one corner": {ptRect(80, 80, 5, 5), 1},
		"touches the origin": {ptRect(0, 0, 10, 10), 1},
		"far away":           {ptRect(500, 500, 10, 10), 0},
	}
	for name, tc := range regions {
		if got := len(ptHits(p, tc.rect)); got != tc.want {
			t.Fatalf("%s: expected %d hits, found %d", name, tc.want, got)
		}
	}
}

func TestNegativeCoordinatesAndMultiCellBounds(t *testing.T) {
	w := donburi.NewWorld()
	p := newPartitions(ptResolutions)

	entry := ptNewEntry(t, w, ptRect(-48, -48, 40, 40))
	ptMustAdd(t, p, entry)

	if got := ptIndex(t, entry).partition; got != 2 {
		t.Fatalf("expected partition 2 for a 40x40 item spanning negative coordinates, got %d", got)
	}

	region := ptRect(-80, -80, 200, 200)
	if got := ptHitSet(p, region); len(got) != 1 || !containsEntity(got, entry.Entity()) {
		t.Fatalf("expected the item exactly once in the negative world region, got %v", ptHits(p, region))
	}

	probe := ptRect(-10, -10, 5, 5)
	if got := ptHitSet(p, probe); len(got) != 1 || !containsEntity(got, entry.Entity()) {
		t.Fatalf("expected the item to be found near the origin, got %v", ptHits(p, probe))
	}
}

func TestTransformMatrixAffectsIndexBounds(t *testing.T) {
	w := donburi.NewWorld()
	p := newPartitions(ptResolutions)
	entry := ptNewEntry(t, w, ptRect(10, 10, 12, 8))

	tr, ok := transform.GetTransform(entry)
	if !ok {
		t.Fatal("entry missing transform")
	}
	tr.SetPosition(100, -30)
	tr.SetScale(2, 2)
	ptMustAdd(t, p, entry)

	b, ok := transform.GetBounds(entry)
	if !ok {
		t.Fatal("entry missing bounds")
	}

	want := b.WorldRect(tr)
	if got := p.getBounds(entry); got != want {
		t.Fatalf("expected transformed bounds to be used for partitioning, got %+v want %+v", got, want)
	}

	if got := ptIndex(t, entry).partition; got != 1 {
		t.Fatalf("expected the world-space size 24x16 to select partition 1, got %d", got)
	}

	query := ptRect(80, -50, 80, 80)
	if got := ptHitSet(p, query); len(got) != 1 || !containsEntity(got, entry.Entity()) {
		t.Fatalf("expected transformed item to be found in its world-space region, got %v", ptHits(p, query))
	}
}

func TestExactPartitionBoundaryEdges(t *testing.T) {
	p := newPartitions(ptResolutions)

	cases := []struct {
		name string
		rect geom.Rect
		want int
	}{
		{"exact_fine_boundary", ptRect(0, 0, 16, 16), 0},
		{"just_over_fine_boundary", ptRect(0, 0, 16.01, 4), 1},
		{"exact_middle_boundary", ptRect(0, 0, 32, 32), 1},
		{"just_over_middle_boundary", ptRect(0, 0, 32.01, 32.01), 2},
		{"exact_coarse_boundary", ptRect(0, 0, 64, 64), 2},
		{"larger_than_coarse", ptRect(0, 0, 128, 12), 2},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := p.nearestFit(tc.rect); got != tc.want {
				t.Fatalf("nearestFit(%+v) = %d, want %d", tc.rect, got, tc.want)
			}
		})
	}
}

func TestRemove(t *testing.T) {
	w := donburi.NewWorld()
	p := newPartitions(ptResolutions)
	entry := ptNewEntry(t, w, ptRect(0, 0, 12, 10))
	ptMustAdd(t, p, entry)

	if err := p.remove(entry); err != nil {
		t.Fatalf("remove failed: %v", err)
	}

	idx := ptIndex(t, entry)
	if idx.partition != -1 || idx.key != 0 {
		t.Fatalf("expected a cleared index, got partition=%d key=%v", idx.partition, idx.key)
	}
	if got := ptCounts(p); !slices.Equal(got, []int{0, 0, 0}) {
		t.Fatalf("expected empty grids, got %v", got)
	}
	if hits := ptHits(p, ptRect(0, 0, 256, 256)); len(hits) != 0 {
		t.Fatalf("expected no hits after remove, got %v", hits)
	}

	if err := p.remove(entry); err != nil {
		t.Fatalf("second remove should be a no-op, got: %v", err)
	}
}

func TestRemoveThenAddAgain(t *testing.T) {
	w := donburi.NewWorld()
	p := newPartitions(ptResolutions)
	entry := ptNewEntry(t, w, ptRect(0, 0, 12, 10))

	ptMustAdd(t, p, entry)
	if err := p.remove(entry); err != nil {
		t.Fatalf("remove failed: %v", err)
	}
	ptMustAdd(t, p, entry)

	idx := ptIndex(t, entry)
	if idx.partition != 0 || idx.key == 0 {
		t.Fatalf("expected the entry to be indexed again, got partition=%d key=%v", idx.partition, idx.key)
	}
	if got := ptCounts(p); !slices.Equal(got, []int{1, 0, 0}) {
		t.Fatalf("unexpected item counts: %v", got)
	}
}

func TestStaleKeyRecovery(t *testing.T) {
	w := donburi.NewWorld()
	p := newPartitions(ptResolutions)
	entry := ptNewEntry(t, w, ptRect(0, 0, 12, 10))
	ptMustAdd(t, p, entry)

	idx := ptIndex(t, entry)
	stale := idx.key
	if err := p.grids[idx.partition].Delete(stale); err != nil {
		t.Fatalf("direct delete failed: %v", err)
	}

	ptMustUpdate(t, p, entry)

	idx = ptIndex(t, entry)
	if idx.partition != 0 || idx.key == 0 || idx.key == stale {
		t.Fatalf("expected a fresh key in partition 0, got partition=%d key=%v (stale=%v)", idx.partition, idx.key, stale)
	}
	if !p.grids[0].Has(idx.key) {
		t.Fatal("new key is not in the grid")
	}
	if got := ptCounts(p); !slices.Equal(got, []int{1, 0, 0}) {
		t.Fatalf("unexpected item counts: %v", got)
	}
}

func TestEntryWithoutIndexingComponentIsIgnored(t *testing.T) {
	w := donburi.NewWorld()
	p := newPartitions(ptResolutions)
	entry := w.Entry(w.Create(transform.TransformComponent))

	if err := p.add(entry); err != nil {
		t.Fatalf("add should ignore it, got: %v", err)
	}
	if err := p.update(entry); err != nil {
		t.Fatalf("update should ignore it, got: %v", err)
	}
	if err := p.remove(entry); err != nil {
		t.Fatalf("remove should ignore it, got: %v", err)
	}
	if got := ptCounts(p); !slices.Equal(got, []int{0, 0, 0}) {
		t.Fatalf("expected empty grids, got %v", got)
	}
}

func TestDontPartitionEntryIsNotIndexed(t *testing.T) {
	w := donburi.NewWorld()
	p := newPartitions(ptResolutions)
	entry := ptNewEntry(t, w, ptRect(0, 0, 12, 10), DontPartition)

	ptMustAdd(t, p, entry)
	ptMustUpdate(t, p, entry)

	idx := ptIndex(t, entry)
	if idx.partition != -1 || idx.key != 0 {
		t.Fatalf("expected it to stay unindexed, got partition=%d key=%v", idx.partition, idx.key)
	}
	if got := ptCounts(p); !slices.Equal(got, []int{0, 0, 0}) {
		t.Fatalf("expected empty grids, got %v", got)
	}
}

func TestWorldQueryAndRangeOver(t *testing.T) {
	w := NewWorld()

	inA := w.Create()
	inB := w.Create()
	out := w.Create()

	for _, tc := range []struct {
		entry *donburi.Entry
		rect  geom.Rect
	}{
		{inA, ptRect(0, 0, 10, 10)},
		{inB, ptRect(25, 10, 8, 12)},
		{out, ptRect(200, 200, 10, 10)},
	} {
		b, bOk := transform.GetBounds(tc.entry)
		if !bOk {
			t.Fatalf("expected entry to have bounds component")
		}

		b.SetRect(tc.rect)
		if err := w.Add(tc.entry); err != nil {
			t.Fatalf("add failed for %+v: %v", tc.rect, err)
		}
	}

	region := ptRect(0, 0, 64, 64)
	got := w.Query(region, nil)
	if len(got) != 2 {
		t.Fatalf("Query(region) = %d entries, want 2", len(got))
	}
	seen := map[donburi.Entity]struct{}{}
	for _, e := range got {
		seen[e.Entity()] = struct{}{}
	}
	if !containsEntity(seen, inA.Entity()) || !containsEntity(seen, inB.Entity()) {
		t.Fatalf("Query(region) missing expected entries: %#v", seen)
	}
	if containsEntity(seen, out.Entity()) {
		t.Fatalf("Query(region) included an out-of-range entity")
	}

	prefix := []*donburi.Entry{out}
	got = w.Query(region, prefix)
	if len(got) != 3 {
		t.Fatalf("Query(region, dst) appended incorrectly: got %d, want 3", len(got))
	}
	if !containsEntity(ptHitSet(w.(*wrld).partitions, region), inA.Entity()) || !containsEntity(ptHitSet(w.(*wrld).partitions, region), inB.Entity()) {
		t.Fatalf("unexpected query result set after append: %#v", got)
	}

	count := 0
	w.RangeOver(region, func(*donburi.Entry) bool {
		count++
		return false // stop after the first
	})
	if count != 1 {
		t.Fatalf("RangeOver did not stop early: count=%d, want 1", count)
	}

	count = 0
	w.RangeOver(region, func(*donburi.Entry) bool {
		count++
		return true
	})
	if count != 2 {
		t.Fatalf("RangeOver should visit exactly 2 entities in the region, got %d", count)
	}
}

// ---------------------------------------------------------------------------
// Randomized consistency
// ---------------------------------------------------------------------------

// TestPartitionsStayConsistentUnderChurn applies thousands of random adds, moves,
// resizes and removes, then checks the whole structure against a simple model:
//   - the grids hold exactly the indexed entities
//   - each indexed entity is in the partition nearestFit chooses, with a live key
//   - a whole-world query finds each indexed entity exactly once and nothing else
//   - each indexed entity is found by a query of its own bounds
//   - the number of keys stored in cells equals the cells the items should cover,
//     which detects residue left behind in old cells
func TestPartitionsStayConsistentUnderChurn(t *testing.T) {
	const (
		entities = 150
		steps    = 4000
		extent   = 2000.0
		sweepLo  = -1000.0
		sweepHi  = 3500.0
	)

	w := donburi.NewWorld()
	p := newPartitions(ptResolutions)
	rng := rand.New(rand.NewPCG(42, 42))

	entries := make([]*donburi.Entry, entities)
	pos := make([][2]float64, entities)
	indexed := make([]bool, entities)
	for i := range entries {
		entries[i] = ptNewEntry(t, w, ptRect(0, 0, 10, 10))
	}

	verify := func(step int) {
		t.Helper()

		hits := map[donburi.Entity]int{}
		world := ptRect(sweepLo, sweepLo, sweepHi-sweepLo, sweepHi-sweepLo)
		for _, e := range ptHits(p, world) {
			hits[e]++
		}

		wantItems := 0
		wantKeys := make([]int, len(p.grids))

		for i, entry := range entries {
			idx := ptIndex(t, entry)
			n := hits[entry.Entity()]

			if !indexed[i] {
				if idx.partition != -1 || idx.key != 0 {
					t.Fatalf("step %d: entity %d should be unindexed, got partition=%d key=%v", step, i, idx.partition, idx.key)
				}
				if n != 0 {
					t.Fatalf("step %d: unindexed entity %d was found %d times", step, i, n)
				}
				continue
			}

			wantItems++
			bounds := p.getBounds(entry)

			if want := p.nearestFit(bounds); idx.partition != want {
				t.Fatalf("step %d: entity %d is in partition %d, expected %d", step, i, idx.partition, want)
			}
			g := p.grids[idx.partition]
			if !g.Has(idx.key) {
				t.Fatalf("step %d: entity %d has a dead key %v", step, i, idx.key)
			}
			if n != 1 {
				t.Fatalf("step %d: entity %d was found %d times by a whole-world query", step, i, n)
			}
			if !slices.Contains(ptHits(p, bounds), entry.Entity()) {
				t.Fatalf("step %d: entity %d was not found by a query of its own bounds %+v", step, i, bounds)
			}

			minX, minY := bounds.Min()
			maxX, maxY := bounds.Max()
			cx1, cy1, cx2, cy2 := g.PaddedRegion(minX, minY, maxX, maxY)
			wantKeys[idx.partition] += int(cx2-cx1) * int(cy2-cy1)
		}

		total := 0
		for _, c := range ptCounts(p) {
			total += c
		}
		if total != wantItems {
			t.Fatalf("step %d: grids hold %d items, expected %d", step, total, wantItems)
		}

		for i, g := range p.grids {
			if got := ptCellKeyTotal(g, sweepLo, sweepHi); got != wantKeys[i] {
				t.Fatalf("step %d: grid %d stores %d cell entries, expected %d (residue or missing cells)", step, i, got, wantKeys[i])
			}
		}
	}

	for step := 1; step <= steps; step++ {
		i := rng.IntN(entities)
		entry := entries[i]

		switch {
		case rng.IntN(5) == 0: // remove
			if err := p.remove(entry); err != nil {
				t.Fatalf("step %d: remove failed: %v", step, err)
			}
			indexed[i] = false

		case indexed[i] && rng.IntN(3) == 0: // small drift, same size
			pos[i][0] += rng.Float64()*6 - 3
			pos[i][1] += rng.Float64()*6 - 3
			ptMove(entry, pos[i][0], pos[i][1])
			ptMustUpdate(t, p, entry)

		default: // jump and resize, possibly across partitions
			size := 4 + rng.Float64()*96
			b, bOk := transform.GetBounds(entry)
			if !bOk {
				t.Fatalf("expected entry to have bounds component")
			}

			b.SetRect(ptRect(0, 0, size, size))
			pos[i] = [2]float64{rng.Float64() * extent, rng.Float64() * extent}
			ptMove(entry, pos[i][0], pos[i][1])
			ptMustUpdate(t, p, entry)
			indexed[i] = true
		}

		if step%100 == 0 {
			verify(step)
		}
	}
	verify(steps)
}
