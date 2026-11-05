package slotmap_test

import (
	"testing"

	"github.com/adm87/finch/pkg/slotmap"
)

func TestInsertGet(t *testing.T) {
	s := slotmap.New[string](2)

	a := s.Insert("a")
	b := s.Insert("b")

	for k, want := range map[slotmap.Key]string{a: "a", b: "b"} {
		got, ok := s.Get(k)
		if !ok || got != want {
			t.Errorf("Get(%v) = (%q, %v), want (%q, true)", k, got, ok, want)
		}
	}
	if s.Len() != 2 {
		t.Errorf("Len() = %d, want 2", s.Len())
	}
}

func TestZeroValue(t *testing.T) {
	var s slotmap.SlotMap[int]
	if s.Len() != 0 || s.Cap() != 0 {
		t.Fatalf("Len, Cap = %d, %d, want 0, 0", s.Len(), s.Cap())
	}
	if _, ok := s.Get(slotmap.Key{}); ok {
		t.Error("Get on empty map ok = true, want false")
	}

	k := s.Insert(1)
	if got, ok := s.Get(k); !ok || got != 1 {
		t.Errorf("Get = (%d, %v), want (1, true)", got, ok)
	}
	if index(k) == 0 {
		t.Error("Insert handed out reserved slot 0")
	}
}

func TestInvalidKey(t *testing.T) {
	s := slotmap.New[int](1)

	for _, k := range []slotmap.Key{{}, slotmap.Unpack(1<<32 | 5)} {
		if _, ok := s.Get(k); ok {
			t.Errorf("Get(%v) ok = true, want false", k)
		}
	}
}

func TestDeleteStalesKey(t *testing.T) {
	s := slotmap.New[int](1)
	k := s.Insert(7)

	got, ok := s.Delete(k)
	if !ok || got != 7 {
		t.Fatalf("Delete = (%d, %v), want (7, true)", got, ok)
	}
	if _, ok := s.Get(k); ok {
		t.Error("Get after Delete ok = true, want false")
	}
	if _, ok := s.Delete(k); ok {
		t.Error("second Delete ok = true, want false")
	}
	if s.Set(k, 1) {
		t.Error("Set after Delete = true, want false")
	}

	k2 := s.Insert(8)
	if index(k2) != index(k) {
		t.Errorf("reinsert index = %d, want reused %d", index(k2), index(k))
	}
	if k2 == k {
		t.Error("reinsert reused the old version")
	}
	if _, ok := s.Get(k); ok {
		t.Error("stale Get after reuse ok = true, want false")
	}
	if s.Len() != 1 {
		t.Errorf("Len() = %d, want 1", s.Len())
	}
}

func TestSetAndGetPtr(t *testing.T) {
	s := slotmap.New[int](0)
	k := s.Insert(1)

	if !s.Set(k, 2) {
		t.Fatal("Set = false, want true")
	}
	p, ok := s.GetPtr(k)
	if !ok {
		t.Fatal("GetPtr ok = false, want true")
	}
	*p = 3

	if got, _ := s.Get(k); got != 3 {
		t.Errorf("Get = %d, want 3", got)
	}
}

func TestGrowKeepsKeys(t *testing.T) {
	s := slotmap.New[int](0)
	if s.Cap() != 0 {
		t.Fatalf("Cap() = %d, want 0", s.Cap())
	}

	keys := make([]slotmap.Key, 100)
	for i := range keys {
		keys[i] = s.Insert(i)
	}
	for i, k := range keys {
		if got, ok := s.Get(k); !ok || got != i {
			t.Errorf("Get(keys[%d]) = (%d, %v), want (%d, true)", i, got, ok, i)
		}
	}
	if s.Cap() < 100 {
		t.Errorf("Cap() = %d, want at least 100", s.Cap())
	}
}

func TestGrowWithFreeSlots(t *testing.T) {
	s := slotmap.New[int](2)
	k := s.Insert(1)
	s.Delete(k)
	s.Grow(2)

	seen := map[uint32]bool{}
	for range 4 {
		seen[index(s.Insert(0))] = true
	}
	if len(seen) != 4 || s.Cap() != 4 {
		t.Errorf("distinct indices = %d, Cap() = %d, want 4 and 4", len(seen), s.Cap())
	}
}

func TestAll(t *testing.T) {
	s := slotmap.New[int](4)
	for i := range 4 {
		s.Insert(i)
	}
	s.Delete(s.Insert(99))

	sum := 0
	for k, v := range s.All() {
		if got, _ := s.Get(k); got != v {
			t.Errorf("All key %v gives %d, Get gives %d", k, v, got)
		}
		sum += v
	}
	if sum != 0+1+2+3 {
		t.Errorf("All sum = %d, want 6", sum)
	}

	calls := 0
	for range s.All() {
		calls++
		break
	}
	if calls != 1 {
		t.Errorf("All after break made %d calls, want 1", calls)
	}
}

func TestDeleteKeepsOtherKeys(t *testing.T) {
	s := slotmap.New[string](3)
	a := s.Insert("a")
	b := s.Insert("b")
	c := s.Insert("c")

	s.Delete(a)
	for k, want := range map[slotmap.Key]string{b: "b", c: "c"} {
		if got, ok := s.Get(k); !ok || got != want {
			t.Errorf("Get(%v) = (%q, %v), want (%q, true)", k, got, ok, want)
		}
	}
}

func TestDeleteDuringAll(t *testing.T) {
	s := slotmap.New[int](5)
	for i := range 5 {
		s.Insert(i)
	}

	seen := map[int]bool{}
	for k, v := range s.All() {
		if seen[v] {
			t.Errorf("All yielded %d twice", v)
		}
		seen[v] = true
		s.Delete(k)
	}
	if len(seen) != 5 || s.Len() != 0 {
		t.Errorf("visited %d, Len() = %d, want 5 and 0", len(seen), s.Len())
	}
}

func TestAllSeesSetAfterGrowth(t *testing.T) {
	s := slotmap.New[int](3)
	a := s.Insert(1)
	s.Insert(2)
	s.Insert(3)

	grown := false
	var got int
	for k, v := range s.All() {
		if !grown {
			s.Insert(4) // full, so this reallocates
			s.Set(a, 100)
			grown = true
		}
		if k == a {
			got = v
		}
	}
	if got != 100 {
		t.Errorf("All yielded %d for a, want 100 set after growth", got)
	}
}

func TestPackRoundTrip(t *testing.T) {
	s := slotmap.New[int](1)
	k := s.Insert(1)

	if got := slotmap.Unpack(k.Pack()); got != k {
		t.Errorf("Unpack(Pack(%v)) = %v", k, got)
	}
}

func TestGrowPastMaxPanics(t *testing.T) {
	defer func() {
		if recover() == nil {
			t.Error("Grow past the slot limit did not panic")
		}
	}()
	s := slotmap.New[struct{}](0)
	s.Grow(int(^uint(0) >> 1))
}

// The slot index is the low half of a packed key.
func index(k slotmap.Key) uint32 {
	return uint32(k.Pack())
}
