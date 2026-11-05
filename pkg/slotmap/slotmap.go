// Package slotmap stores values under keys that stop working once the value is deleted.
package slotmap

import (
	"iter"
	"math"
	"slices"
)

// Every slot index has to fit in a uint32, and every slice length in an int.
const maxSlots = min(math.MaxUint32, math.MaxInt)

type slot struct {
	pos     uint32 // index into values while occupied
	next    uint32 // next free slot while free
	version uint32
}

// The low bit of the version is set while the slot holds a value.
func (s *slot) occupied() bool {
	return s.version&1 == 1
}

// Key finds a value in a SlotMap. The zero Key is never valid.
type Key struct {
	index   uint32
	version uint32
}

// Unpack turns a packed uint64 back into a Key.
func Unpack(packed uint64) Key {
	return Key{
		index:   uint32(packed),
		version: uint32(packed >> 32),
	}
}

// Pack fits k into a uint64 so it can be stored elsewhere.
func (k Key) Pack() uint64 {
	return uint64(k.version)<<32 | uint64(k.index)
}

// SlotMap holds values under keys that stop working once the value is deleted.
// Its zero value is ready to use.
type SlotMap[T any] struct {
	slots   []slot
	values  []T      // packed so All only visits live values
	slotOf  []uint32 // slot index of each value
	head    uint32   // first free slot; 0 means none, so slot 0 is never used
	retired int      // slots whose version ran out
}

// New makes room for capacity values up front.
func New[T any](capacity int) *SlotMap[T] {
	s := &SlotMap[T]{}
	s.Grow(capacity)
	return s
}

func (s *SlotMap[T]) Len() int {
	return len(s.values)
}

func (s *SlotMap[T]) Cap() int {
	return max(len(s.slots)-1-s.retired, 0)
}

func (s *SlotMap[T]) Insert(value T) Key {
	if s.head == 0 {
		s.Grow(growBy(len(s.slots)))
	}

	index := s.head
	sl := &s.slots[index]
	s.head = sl.next

	sl.version++
	sl.pos = uint32(len(s.values))
	s.values = append(s.values, value)
	s.slotOf = append(s.slotOf, index)
	return Key{index: index, version: sl.version}
}

// Get reports false once k's value is deleted.
func (s *SlotMap[T]) Get(k Key) (T, bool) {
	sl, ok := s.lookup(k)
	if !ok {
		var zero T
		return zero, false
	}
	return s.values[sl.pos], true
}

// GetPtr returns a pointer to the value, or nil if k's value is deleted.
// The pointer is only valid until the next Insert, Delete or Grow.
func (s *SlotMap[T]) GetPtr(k Key) (*T, bool) {
	sl, ok := s.lookup(k)
	if !ok {
		return nil, false
	}
	return &s.values[sl.pos], true
}

func (s *SlotMap[T]) Set(k Key, value T) bool {
	sl, ok := s.lookup(k)
	if !ok {
		return false
	}
	s.values[sl.pos] = value
	return true
}

func (s *SlotMap[T]) Delete(k Key) (T, bool) {
	var zero T

	sl, ok := s.lookup(k)
	if !ok {
		return zero, false
	}
	old := s.values[sl.pos]

	// Move the last value into the hole so values stay packed.
	last := len(s.values) - 1
	s.values[sl.pos] = s.values[last]
	s.slotOf[sl.pos] = s.slotOf[last]
	s.slots[s.slotOf[sl.pos]].pos = sl.pos
	s.values[last] = zero
	s.values = s.values[:last]
	s.slotOf = s.slotOf[:last]

	sl.version++
	if sl.version == math.MaxUint32-1 {
		// One more use would wrap the version and bring old keys back.
		s.retired++
		return old, true
	}
	sl.next = s.head
	s.head = k.index
	return old, true
}

// All yields in no particular order. Deleting the current key during All is safe.
func (s *SlotMap[T]) All() iter.Seq2[Key, T] {
	return func(yield func(Key, T) bool) {
		// Iterate backwards so that deleting the current key doesn't affect the next iteration.
		// Backward holds the slice it started with, so read values live in case an Insert grew it.
		for i := range slices.Backward(s.values) {
			if i >= len(s.values) {
				continue
			}
			index := s.slotOf[i]
			if !yield(Key{index: index, version: s.slots[index].version}, s.values[i]) {
				return
			}
		}
	}
}

// Grow adds room for n more values.
func (s *SlotMap[T]) Grow(n int) {
	if n <= 0 {
		return
	}
	if len(s.slots) == 0 {
		s.slots = append(s.slots, slot{})
	}
	if n > maxSlots-len(s.slots) {
		panic("slotmap: too many slots")
	}

	oldLen := uint32(len(s.slots))
	s.slots = append(s.slots, make([]slot, n)...)
	newTail := uint32(len(s.slots) - 1)

	for i := oldLen; i < newTail; i++ {
		s.slots[i].next = i + 1
	}

	// Put the new free slots in front of the old ones.
	s.slots[newTail].next = s.head
	s.head = oldLen

	s.values = slices.Grow(s.values, s.Cap()-len(s.values))
	s.slotOf = slices.Grow(s.slotOf, s.Cap()-len(s.slotOf))
}

func (s *SlotMap[T]) lookup(k Key) (*slot, bool) {
	if k.index == 0 || k.index >= uint32(len(s.slots)) {
		return nil, false
	}

	sl := &s.slots[k.index]
	if sl.version != k.version || !sl.occupied() {
		return nil, false
	}
	return sl, true
}

// growBy doubles the slots, but never past maxSlots.
func growBy(slots int) int {
	return max(min(slots, maxSlots-slots), 1)
}
