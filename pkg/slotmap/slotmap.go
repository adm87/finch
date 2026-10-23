package slotmap

import "errors"

var (
	ErrInvalidKey = errors.New("invalid key")
	ErrNotFound   = errors.New("not found")
)

// K is a versioned slot key. The version changes each time a slot is reused,
// so a stale key can't reach a newer value.
type K struct {
	index   uint32
	version uint32
}

func (k K) Index() uint32 {
	return k.index
}

func (k K) Version() uint32 {
	return k.version
}

func (k K) Pack() uint64 {
	return uint64(k.version)<<32 | uint64(k.index)
}

func Unpack(packed uint64) K {
	return K{
		index:   uint32(packed & 0xFFFFFFFF),
		version: uint32(packed >> 32),
	}
}

// checkCapacity panics if capacity is negative or doesn't fit a uint32 index.
func checkCapacity(newLen int) {
	if newLen < 0 {
		panic("slotmap: capacity is negative")
	}
	if uint64(newLen) > uint64(^uint32(0)) {
		panic("slotmap: capacity exceeds uint32 index space")
	}
	maxInt := int(^uint(0) >> 1)
	if newLen >= maxInt {
		panic("slotmap: capacity too large")
	}
}

type slotData[T any] struct {
	value T
	next  uint32
}

type slot[T any] struct {
	data    slotData[T]
	version uint32
}

// IsOccupied reads the low bit of the version, which is set while the slot holds a value.
func (s slot[T]) IsOccupied() bool {
	return s.version&1 == 1
}

// SlotMap stores values under versioned keys, with O(1) insert and delete.
type SlotMap[T any] struct {
	slots []slot[T]
	head  uint32
	len   int
}

// New creates a SlotMap that holds capacity values before it grows.
func New[T any](capacity int) *SlotMap[T] {
	capacity = max(capacity, 0)
	checkCapacity(capacity + 1)

	slotMap := &SlotMap[T]{
		slots: make([]slot[T], capacity+1),
		head:  0,
	}

	if capacity == 0 {
		return slotMap
	}

	for i := 1; i < len(slotMap.slots); i++ {
		slotMap.slots[i].data.next = uint32(i + 1)
	}
	slotMap.slots[len(slotMap.slots)-1].data.next = 0
	slotMap.head = 1

	return slotMap
}

// Insert adds value and returns its key, growing the map if it is full.
func (s *SlotMap[T]) Insert(value T) K {
	if s.head == 0 {
		s.Grow(max(1, len(s.slots)))
	}

	index := s.head

	slot := &s.slots[index]
	s.head = slot.data.next

	slot.version++
	slot.data.value = value

	s.len++
	return K{
		index:   index,
		version: slot.version,
	}
}

// Get returns ErrInvalidKey for a malformed key and ErrNotFound for a deleted one.
func (s *SlotMap[T]) Get(k K) (T, error) {
	var zero T

	index, version := k.index, k.version
	if index == 0 || index >= uint32(len(s.slots)) {
		return zero, ErrInvalidKey
	}

	slot := &s.slots[index]
	if slot.version != version || !slot.IsOccupied() {
		return zero, ErrNotFound
	}

	return slot.data.value, nil
}

// GetPtr is like Get but returns a pointer to the stored value.
func (s *SlotMap[T]) GetPtr(k K) (*T, error) {
	index, version := k.index, k.version
	if index == 0 || index >= uint32(len(s.slots)) {
		return nil, ErrInvalidKey
	}

	slot := &s.slots[index]
	if slot.version != version || !slot.IsOccupied() {
		return nil, ErrNotFound
	}

	return &slot.data.value, nil
}

func (s *SlotMap[T]) Set(k K, value T) error {
	index, version := k.index, k.version
	if index == 0 || index >= uint32(len(s.slots)) {
		return ErrInvalidKey
	}

	slot := &s.slots[index]
	if slot.version != version || !slot.IsOccupied() {
		return ErrNotFound
	}

	slot.data.value = value
	return nil
}

// Delete removes and returns the value at k. The key stays invalid afterward.
func (s *SlotMap[T]) Delete(k K) (T, error) {
	var zero T

	index, version := k.index, k.version
	if index == 0 || index >= uint32(len(s.slots)) {
		return zero, ErrInvalidKey
	}

	slot := &s.slots[index]
	if slot.version != version || !slot.IsOccupied() {
		return zero, ErrNotFound
	}

	oldValue := slot.data.value
	slot.data.value = zero

	slot.version++
	slot.data.next = s.head

	s.head = index
	s.len--
	return oldValue, nil
}

// Range calls fn for each value until fn returns false. The order is not stable.
func (s *SlotMap[T]) Range(fn func(k K, value T) bool) {
	for i := 1; i < len(s.slots); i++ {
		slot := &s.slots[i]
		if slot.IsOccupied() {
			k := K{
				index:   uint32(i),
				version: slot.version,
			}
			if !fn(k, slot.data.value) {
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

	newLen := len(s.slots) + n
	checkCapacity(newLen)

	oldLen := uint32(len(s.slots))
	s.slots = append(s.slots, make([]slot[T], n)...)
	newTail := uint32(len(s.slots) - 1)

	for i := oldLen; i < newTail; i++ {
		s.slots[i].data.next = i + 1
	}

	// Chain the new free-list segment onto the existing head.
	s.slots[newTail].data.next = s.head
	s.head = oldLen
}

func (s *SlotMap[T]) Len() int {
	return s.len
}

// Cap returns how many values fit before the map grows.
func (s *SlotMap[T]) Cap() int {
	return len(s.slots) - 1 // subtract 1 for the sentinel slot at index 0
}
