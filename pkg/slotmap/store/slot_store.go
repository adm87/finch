package store

import (
	"github.com/adm87/finch/pkg/errs"
	"github.com/adm87/finch/pkg/slotmap"
)

type options struct {
	kind     string
	capacity int
}

type Option func(*options)

func WithCapacity(capacity int) Option {
	return func(o *options) {
		o.capacity = capacity
	}
}

// WithKind names what the storage holds, for error messages.
func WithKind(kind string) Option {
	return func(o *options) {
		o.kind = kind
	}
}

type Storage[H ~uint64, T any] struct {
	store        *slotmap.SlotMap[T]
	handleByName map[string]H
	nameByHandle map[H]string
	kind         string
}

func New[H ~uint64, T any](opts ...Option) *Storage[H, T] {
	options := &options{
		kind:     "slot",
		capacity: 128,
	}
	for _, opt := range opts {
		opt(options)
	}
	return &Storage[H, T]{
		store:        slotmap.New[T](options.capacity),
		handleByName: make(map[string]H),
		nameByHandle: make(map[H]string),
		kind:         options.kind,
	}
}

func (s *Storage[H, T]) Lookup(name string) (H, bool) {
	handle, ok := s.handleByName[name]
	return handle, ok
}

func (s *Storage[H, T]) Insert(name string, value T) (H, error) {
	if _, exists := s.handleByName[name]; exists {
		return 0, errs.AlreadyExists(s.kind, name)
	}

	handle := H(s.store.Insert(value).Pack())

	s.handleByName[name] = handle
	s.nameByHandle[handle] = name

	return handle, nil
}

func (s *Storage[H, T]) Get(handle H) (T, error) {
	value, err := s.store.Get(slotmap.Unpack(uint64(handle)))

	if err != nil {
		var zero T
		return zero, slotmap.HandleError("get", s.kind, uint64(handle), err)
	}

	return value, nil
}

func (s *Storage[H, T]) Delete(handle H) (T, error) {
	name, found := s.nameByHandle[handle]
	if !found {
		var zero T
		return zero, slotmap.HandleError("delete", s.kind, uint64(handle), slotmap.ErrNotFound)
	}

	value, err := s.store.Delete(slotmap.Unpack(uint64(handle)))
	if err != nil {
		var zero T
		return zero, slotmap.HandleError("delete", s.kind, uint64(handle), err)
	}

	delete(s.handleByName, name)
	delete(s.nameByHandle, handle)

	return value, nil
}
