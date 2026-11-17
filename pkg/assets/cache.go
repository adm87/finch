package assets

import (
	"errors"
	"fmt"
	"io/fs"

	"github.com/adm87/finch/pkg/slotmap"
)

var ErrNotLoaded = errors.New("assets: not loaded")

// Handle finds a loaded asset in its Cache. The zero Handle is never valid.
type Handle[T any] uint64

// Importer turns the file at path into a T.
type Importer[T any] interface {
	Import(fsys fs.FS, path string) (T, error)
}

type ImporterFunc[T any] func(fsys fs.FS, path string) (T, error)

func (f ImporterFunc[T]) Import(fsys fs.FS, path string) (T, error) {
	return f(fsys, path)
}

// Releaser frees whatever a T holds once nothing uses it.
type Releaser[T any] interface {
	Release(value T)
}

type ReleaserFunc[T any] func(value T)

func (f ReleaserFunc[T]) Release(value T) {
	f(value)
}

type entry[T any] struct {
	value T
	refs  int
}

// Cache loads each path once and counts how many times it was loaded.
type Cache[T any] struct {
	fsys     fs.FS
	importer Importer[T]
	releaser Releaser[T]
	handles  map[string]Handle[T]
	entries  slotmap.SlotMap[entry[T]]
}

// NewCache imports from fsys. releaser can be nil if T needs no cleanup.
func NewCache[T any](fsys fs.FS, importer Importer[T], releaser Releaser[T]) *Cache[T] {
	if fsys == nil {
		panic("assets: fsys is nil")
	}
	if importer == nil {
		panic("assets: importer is nil")
	}
	return &Cache[T]{
		fsys:     fsys,
		importer: importer,
		releaser: releaser,
		handles:  make(map[string]Handle[T]),
	}
}

func (c *Cache[T]) Len() int {
	return c.entries.Len()
}

// Load imports path the first time, and adds a reference every time.
func (c *Cache[T]) Load(path string) (Handle[T], error) {
	if h, ok := c.handles[path]; ok {
		e, _ := c.entries.GetPtr(slotmap.Unpack(uint64(h)))
		e.refs++
		return h, nil
	}

	value, err := c.importer.Import(c.fsys, path)
	if err != nil {
		return 0, fmt.Errorf("load %q: %w", path, err)
	}

	h := Handle[T](c.entries.Insert(entry[T]{value: value, refs: 1}).Pack())
	c.handles[path] = h
	return h, nil
}

// Unload drops a reference, and releases path once none are left.
func (c *Cache[T]) Unload(path string) error {
	h, ok := c.handles[path]
	if !ok {
		return fmt.Errorf("unload %q: %w", path, ErrNotLoaded)
	}

	key := slotmap.Unpack(uint64(h))

	e, _ := c.entries.GetPtr(key)
	e.refs--

	if e.refs > 0 {
		return nil
	}

	old, _ := c.entries.Delete(key)
	delete(c.handles, path)

	c.release(old.value)
	return nil
}

// Get returns the asset h points to, or false once it's been unloaded.
func (c *Cache[T]) Get(h Handle[T]) (T, bool) {
	e, ok := c.entries.Get(slotmap.Unpack(uint64(h)))
	return e.value, ok
}

// Lookup finds the handle for a loaded path without adding a reference.
func (c *Cache[T]) Lookup(path string) (Handle[T], bool) {
	h, ok := c.handles[path]
	return h, ok
}

// Clear releases everything, whatever its reference count.
func (c *Cache[T]) Clear() {
	for key, e := range c.entries.All() {
		c.entries.Delete(key)
		c.release(e.value)
	}
	clear(c.handles)
}

func (c *Cache[T]) release(value T) {
	if c.releaser != nil {
		c.releaser.Release(value)
	}
}
