package world

import (
	"errors"
	"testing"

	"github.com/adm87/finch/pkg/geom"
	"github.com/yohamta/donburi"
)

func newObjectWithBounds(world donburi.World, rect geom.Rect) WorldObject {
	obj := NewWorldObject(world)
	obj.Bounds(world).SetRect(rect)
	// Ensure transform generates a predictable identity-like matrix path.
	obj.Transform(world).SetScale(1, 1)
	return obj
}

func TestMakePartitions(t *testing.T) {
	t.Run("panics when resolutions are empty", func(t *testing.T) {
		defer func() {
			if recover() == nil {
				t.Fatal("expected panic for empty resolutions")
			}
		}()
		_ = makePartitions(nil)
	})

	t.Run("panics when resolutions contain non-positive values", func(t *testing.T) {
		defer func() {
			if recover() == nil {
				t.Fatal("expected panic for non-positive resolution")
			}
		}()
		_ = makePartitions([]int{16, 0, 32})
	})

	t.Run("sorts resolutions in ascending order", func(t *testing.T) {
		p := makePartitions([]int{64, 16, 32})
		if len(p.grids) != 3 {
			t.Fatalf("expected 3 partitions, got %d", len(p.grids))
		}

		got := []int{
			p.grids[0].CellSize(),
			p.grids[1].CellSize(),
			p.grids[2].CellSize(),
		}
		want := []int{16, 32, 64}
		for i := range want {
			if got[i] != want[i] {
				t.Fatalf("expected sorted sizes %v, got %v", want, got)
			}
		}
	})
}

func TestPartitionsNearestFit(t *testing.T) {
	p := makePartitions([]int{16, 32, 64})

	cases := []struct {
		name   string
		bounds geom.Rect
		want   int
	}{
		{name: "fits first partition", bounds: geom.NewRect(0, 0, 12, 8), want: 0},
		{name: "fits exact first size", bounds: geom.NewRect(0, 0, 16, 4), want: 0},
		{name: "fits second partition", bounds: geom.NewRect(0, 0, 24, 20), want: 1},
		{name: "fits third partition", bounds: geom.NewRect(0, 0, 40, 30), want: 2},
		{name: "falls back to last partition", bounds: geom.NewRect(0, 0, 90, 70), want: 2},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := p.nearestFit(tc.bounds); got != tc.want {
				t.Fatalf("nearestFit(%+v) = %d, want %d", tc.bounds, got, tc.want)
			}
		})
	}
}

func TestPartitionsAddUpdateRemove(t *testing.T) {
	world := donburi.NewWorld()
	p := makePartitions([]int{16, 32, 64})
	obj := newObjectWithBounds(world, geom.NewRect(0, 0, 12, 10))

	t.Run("add stores object in expected partition", func(t *testing.T) {
		if err := p.add(world, obj); err != nil {
			t.Fatalf("add failed: %v", err)
		}

		idx := index.Get(obj.Entry(world))
		if idx.partition != 0 {
			t.Fatalf("expected partition 0, got %d", idx.partition)
		}
		if idx.key == 0 {
			t.Fatal("expected non-zero key after add")
		}
		if p.grids[0].ItemCount() != 1 || p.grids[1].ItemCount() != 0 || p.grids[2].ItemCount() != 0 {
			t.Fatalf("unexpected item counts after add: [%d %d %d]",
				p.grids[0].ItemCount(), p.grids[1].ItemCount(), p.grids[2].ItemCount())
		}
	})

	t.Run("update inside same partition keeps partition and key", func(t *testing.T) {
		idxBefore := *index.Get(obj.Entry(world))
		obj.Bounds(world).SetRect(geom.NewRect(8, 8, 14, 9))

		if err := p.update(world, obj); err != nil {
			t.Fatalf("update failed: %v", err)
		}

		idxAfter := index.Get(obj.Entry(world))
		if idxAfter.partition != idxBefore.partition {
			t.Fatalf("expected same partition %d, got %d", idxBefore.partition, idxAfter.partition)
		}
		if idxAfter.key != idxBefore.key {
			t.Fatalf("expected same key %v, got %v", idxBefore.key, idxAfter.key)
		}
	})

	t.Run("update to larger bounds moves object to larger partition", func(t *testing.T) {
		obj.Bounds(world).SetRect(geom.NewRect(0, 0, 48, 40))

		if err := p.update(world, obj); err != nil {
			t.Fatalf("cross-partition update failed: %v", err)
		}

		idx := index.Get(obj.Entry(world))
		if idx.partition != 2 {
			t.Fatalf("expected partition 2 after resize, got %d", idx.partition)
		}
		if idx.key == 0 {
			t.Fatal("expected non-zero key after cross-partition update")
		}
		if p.grids[0].ItemCount() != 0 || p.grids[1].ItemCount() != 0 || p.grids[2].ItemCount() != 1 {
			t.Fatalf("unexpected item counts after cross-partition update: [%d %d %d]",
				p.grids[0].ItemCount(), p.grids[1].ItemCount(), p.grids[2].ItemCount())
		}
	})

	t.Run("remove clears index and is idempotent", func(t *testing.T) {
		if err := p.remove(world, obj); err != nil {
			t.Fatalf("remove failed: %v", err)
		}

		idx := index.Get(obj.Entry(world))
		if idx.partition != -1 || idx.key != 0 {
			t.Fatalf("expected cleared index, got partition=%d key=%v", idx.partition, idx.key)
		}
		if p.grids[0].ItemCount() != 0 || p.grids[1].ItemCount() != 0 || p.grids[2].ItemCount() != 0 {
			t.Fatalf("unexpected item counts after remove: [%d %d %d]",
				p.grids[0].ItemCount(), p.grids[1].ItemCount(), p.grids[2].ItemCount())
		}

		if err := p.remove(world, obj); err != nil {
			t.Fatalf("second remove should be a no-op, got error: %v", err)
		}
	})

	t.Run("unpartitioned objects are not indexed", func(t *testing.T) {
		obj := NewWorldObject(world, Unpartitioned)
		obj.Bounds(world).SetRect(geom.NewRect(0, 0, 12, 10))

		if err := p.add(world, obj); err != nil {
			t.Fatalf("add unpartitioned object failed: %v", err)
		}

		idx := index.Get(obj.Entry(world))
		if idx.partition != -1 || idx.key != 0 {
			t.Fatalf("expected unpartitioned object to stay unindexed, got partition=%d key=%v", idx.partition, idx.key)
		}
		if p.grids[0].ItemCount() != 0 || p.grids[1].ItemCount() != 0 || p.grids[2].ItemCount() != 0 {
			t.Fatalf("expected no indexed objects after unpartitioned add: [%d %d %d]",
				p.grids[0].ItemCount(), p.grids[1].ItemCount(), p.grids[2].ItemCount())
		}
	})

	t.Run("adding unpartitioned tag clears existing index entry on update", func(t *testing.T) {
		obj := newObjectWithBounds(world, geom.NewRect(0, 0, 12, 10))
		if err := p.add(world, obj); err != nil {
			t.Fatalf("setup add failed: %v", err)
		}

		entry := obj.Entry(world)
		entry.AddComponent(Unpartitioned)

		if err := p.update(world, obj); err != nil {
			t.Fatalf("update after tagging unpartitioned failed: %v", err)
		}

		idx := index.Get(entry)
		if idx.partition != -1 || idx.key != 0 {
			t.Fatalf("expected index to be cleared after unpartitioned update, got partition=%d key=%v", idx.partition, idx.key)
		}
		if p.grids[0].ItemCount() != 0 || p.grids[1].ItemCount() != 0 || p.grids[2].ItemCount() != 0 {
			t.Fatalf("expected no indexed objects after unpartitioned update: [%d %d %d]",
				p.grids[0].ItemCount(), p.grids[1].ItemCount(), p.grids[2].ItemCount())
		}
	})

	t.Run("remove still clears stale entries even if object is now unpartitioned", func(t *testing.T) {
		obj := newObjectWithBounds(world, geom.NewRect(0, 0, 12, 10))
		if err := p.add(world, obj); err != nil {
			t.Fatalf("setup add failed: %v", err)
		}

		entry := obj.Entry(world)
		entry.AddComponent(Unpartitioned)

		if err := p.remove(world, obj); err != nil {
			t.Fatalf("remove after tagging unpartitioned failed: %v", err)
		}

		idx := index.Get(entry)
		if idx.partition != -1 || idx.key != 0 {
			t.Fatalf("expected index to be cleared by remove, got partition=%d key=%v", idx.partition, idx.key)
		}
		if p.grids[0].ItemCount() != 0 || p.grids[1].ItemCount() != 0 || p.grids[2].ItemCount() != 0 {
			t.Fatalf("expected no indexed objects after remove: [%d %d %d]",
				p.grids[0].ItemCount(), p.grids[1].ItemCount(), p.grids[2].ItemCount())
		}
	})
}

func TestWorldQuery(t *testing.T) {
	w := New(WithResolutions(16, 32, 64))
	ecs := w.ECS()

	hit, err := w.Create()
	if err != nil {
		t.Fatalf("create hit failed: %v", err)
	}
	hit.Bounds(ecs).SetRect(geom.NewRect(0, 0, 4, 4))
	if err := w.Update(hit); err != nil {
		t.Fatalf("update hit failed: %v", err)
	}

	sameCellMiss, err := w.Create()
	if err != nil {
		t.Fatalf("create sameCellMiss failed: %v", err)
	}
	sameCellMiss.Bounds(ecs).SetRect(geom.NewRect(12, 12, 2, 2))
	if err := w.Update(sameCellMiss); err != nil {
		t.Fatalf("update sameCellMiss failed: %v", err)
	}

	offscreen, err := w.Create()
	if err != nil {
		t.Fatalf("create offscreen failed: %v", err)
	}
	offscreen.Bounds(ecs).SetRect(geom.NewRect(64, 64, 4, 4))
	if err := w.Update(offscreen); err != nil {
		t.Fatalf("update offscreen failed: %v", err)
	}

	region := geom.NewRect(0, 0, 4, 4)
	got := w.Query(region, nil)
	found := map[donburi.Entity]bool{}
	for _, obj := range got {
		found[obj.Entity()] = true
	}
	if len(found) != 2 {
		t.Fatalf("expected 2 objects from overlapping partitions, got %d", len(found))
	}
	if !found[hit.Entity()] || !found[sameCellMiss.Entity()] {
		t.Fatalf("expected entities %v and %v, got %v", hit.Entity(), sameCellMiss.Entity(), found)
	}
	if found[offscreen.Entity()] {
		t.Fatalf("did not expect offscreen entity %v in query result", offscreen.Entity())
	}

	buf := make([]WorldObject, 1, 8)
	buf[0] = offscreen
	got = w.Query(region, buf)
	found = map[donburi.Entity]bool{}
	for _, obj := range got {
		found[obj.Entity()] = true
	}
	if len(found) != 3 {
		t.Fatalf("expected 3 objects with reused buffer (existing + query results), got %d", len(found))
	}
	if !found[hit.Entity()] || !found[sameCellMiss.Entity()] || !found[offscreen.Entity()] {
		t.Fatalf("expected entities %v, %v and prefilled %v; got %v",
			hit.Entity(), sameCellMiss.Entity(), offscreen.Entity(), found)
	}
}

func TestWorldRange(t *testing.T) {
	w := New(WithResolutions(16, 32, 64))
	ecs := w.ECS()

	a, err := w.Create()
	if err != nil {
		t.Fatalf("create a failed: %v", err)
	}
	a.Bounds(ecs).SetRect(geom.NewRect(0, 0, 4, 4))
	if err := w.Update(a); err != nil {
		t.Fatalf("update a failed: %v", err)
	}

	sameCellMiss, err := w.Create()
	if err != nil {
		t.Fatalf("create sameCellMiss failed: %v", err)
	}
	sameCellMiss.Bounds(ecs).SetRect(geom.NewRect(12, 12, 3, 3))
	if err := w.Update(sameCellMiss); err != nil {
		t.Fatalf("update sameCellMiss failed: %v", err)
	}

	miss, err := w.Create()
	if err != nil {
		t.Fatalf("create miss failed: %v", err)
	}
	miss.Bounds(ecs).SetRect(geom.NewRect(64, 64, 4, 4))
	if err := w.Update(miss); err != nil {
		t.Fatalf("update miss failed: %v", err)
	}

	region := geom.NewRect(0, 0, 4, 4)
	found := map[donburi.Entity]bool{}
	for obj := range w.Range(region) {
		found[obj.Entity()] = true
	}

	if len(found) != 2 {
		t.Fatalf("expected 2 ranged hits, got %d", len(found))
	}
	if !found[a.Entity()] || !found[sameCellMiss.Entity()] {
		t.Fatalf("expected entities %v and %v from overlapping partitions", a.Entity(), sameCellMiss.Entity())
	}
	if found[miss.Entity()] {
		t.Fatalf("did not expect entity %v to be in range", miss.Entity())
	}

	count := 0
	for range w.Range(region) {
		count++
		break
	}
	if count != 1 {
		t.Fatalf("expected early-stop range iteration to visit 1 object, got %d", count)
	}
}

func TestWorldCreateManyInvalidCountError(t *testing.T) {
	w := New(WithResolutions(16, 32, 64))
	_, err := w.CreateMany(-1)
	if err == nil {
		t.Fatal("expected error for negative count")
	}
	if !errors.Is(err, ErrCreateManyInvalidCount) {
		t.Fatalf("expected ErrCreateManyInvalidCount, got %v", err)
	}
}

func BenchmarkPartitionsAdd(b *testing.B) {
	world := donburi.NewWorld()
	p := makePartitions([]int{16, 32, 64})
	obj := newObjectWithBounds(world, geom.NewRect(0, 0, 12, 10))

	for i := 0; b.Loop(); i++ {
		size := float64((i % 48) + 1) // [1..48], stays within available partition sizes
		obj.Bounds(world).SetRect(geom.NewRect(float64(i%1024), float64((i*3)%1024), size, size))

		if err := p.add(world, obj); err != nil {
			b.Fatalf("add failed: %v", err)
		}
		if err := p.remove(world, obj); err != nil {
			b.Fatalf("remove during benchmark cleanup failed: %v", err)
		}
	}
}

func BenchmarkPartitionsUpdateSamePartition(b *testing.B) {
	world := donburi.NewWorld()
	p := makePartitions([]int{16, 32, 64})
	obj := newObjectWithBounds(world, geom.NewRect(0, 0, 12, 10))
	if err := p.add(world, obj); err != nil {
		b.Fatalf("setup add failed: %v", err)
	}

	for i := 0; b.Loop(); i++ {
		obj.Bounds(world).SetRect(geom.NewRect(float64(i%32), float64((i*5)%32), 14, 9))
		if err := p.update(world, obj); err != nil {
			b.Fatalf("update failed: %v", err)
		}
	}
}

func BenchmarkPartitionsUpdateByMovementSpeed(b *testing.B) {
	cases := []struct {
		name  string
		stepX float64
		stepY float64
	}{
		{name: "slow", stepX: 0.5, stepY: 0.25},
		{name: "medium", stepX: 4, stepY: 2},
		{name: "fast", stepX: 20, stepY: 10},
	}

	for _, tc := range cases {
		b.Run(tc.name, func(b *testing.B) {
			world := donburi.NewWorld()
			p := makePartitions([]int{16, 32, 64})
			obj := newObjectWithBounds(world, geom.NewRect(0, 0, 12, 10))
			if err := p.add(world, obj); err != nil {
				b.Fatalf("setup add failed: %v", err)
			}

			x, y := 0.0, 0.0
			for b.Loop() {
				x += tc.stepX
				y += tc.stepY
				if x > 2048 {
					x = 0
				}
				if y > 2048 {
					y = 0
				}

				obj.Transform(world).SetPosition(x, y)
				if err := p.update(world, obj); err != nil {
					b.Fatalf("update failed: %v", err)
				}
			}
		})
	}
}

func BenchmarkPartitionsUpdateUnpartitioned(b *testing.B) {
	world := donburi.NewWorld()
	p := makePartitions([]int{16, 32, 64})
	obj := NewWorldObject(world, Unpartitioned)
	obj.Bounds(world).SetRect(geom.NewRect(0, 0, 12, 10))

	for i := 0; b.Loop(); i++ {
		obj.Transform(world).SetPosition(float64(i), float64(i/2))
		if err := p.update(world, obj); err != nil {
			b.Fatalf("update failed: %v", err)
		}
	}
}

func BenchmarkPartitionsUpdateAcrossPartitions(b *testing.B) {
	world := donburi.NewWorld()
	p := makePartitions([]int{16, 32, 64})
	obj := newObjectWithBounds(world, geom.NewRect(0, 0, 12, 10))
	if err := p.add(world, obj); err != nil {
		b.Fatalf("setup add failed: %v", err)
	}

	for i := 0; b.Loop(); i++ {
		if i%2 == 0 {
			obj.Bounds(world).SetRect(geom.NewRect(0, 0, 12, 10))
		} else {
			obj.Bounds(world).SetRect(geom.NewRect(0, 0, 48, 40))
		}
		if err := p.update(world, obj); err != nil {
			b.Fatalf("cross-partition update failed: %v", err)
		}
	}
}

func BenchmarkPartitionsRemove(b *testing.B) {
	world := donburi.NewWorld()
	p := makePartitions([]int{16, 32, 64})
	obj := newObjectWithBounds(world, geom.NewRect(0, 0, 12, 10))

	for b.Loop() {
		if err := p.add(world, obj); err != nil {
			b.Fatalf("add failed: %v", err)
		}
		if err := p.remove(world, obj); err != nil {
			b.Fatalf("remove failed: %v", err)
		}
	}
}
