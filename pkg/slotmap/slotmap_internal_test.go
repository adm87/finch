package slotmap

import (
	"math"
	"testing"
)

func TestGrowByStopsAtMaxSlots(t *testing.T) {
	tests := []struct {
		slots, want int
	}{
		{0, 1},
		{4, 4},
		{maxSlots - 3, 3},
		{maxSlots, 1}, // Grow(1) then panics
	}
	for _, tt := range tests {
		if got := growBy(tt.slots); got != tt.want {
			t.Errorf("growBy(%d) = %d, want %d", tt.slots, got, tt.want)
		}
	}
}

func TestDeleteRetiresWornOutSlot(t *testing.T) {
	s := New[int](1)
	k := s.Insert(1)

	k.version = math.MaxUint32 - 2
	s.slots[k.index].version = k.version
	if _, ok := s.Delete(k); !ok {
		t.Fatal("Delete ok = false, want true")
	}

	if s.Cap() != 0 {
		t.Errorf("Cap() = %d, want 0 after retiring the only slot", s.Cap())
	}
	if k2 := s.Insert(2); k2.index == k.index {
		t.Error("Insert reused a retired slot")
	}
}
