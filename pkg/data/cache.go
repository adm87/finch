package data

import (
	"github.com/adm87/finch/pkg/slotmap/store"
)

// Handle refers to data in a Cache.
type Handle uint64

type Cache struct {
	store *store.Storage[Handle, []byte]
}

func NewCache() *Cache {
	return &Cache{
		store: store.New[Handle, []byte](store.WithKind("data")),
	}
}

func (dc *Cache) Get(handle Handle) ([]byte, error) {
	return dc.store.Get(handle)
}

func (dc *Cache) Put(name string, raw []byte) (Handle, error) {
	return dc.store.Insert(name, raw)
}

func (dc *Cache) Delete(handle Handle) ([]byte, error) {
	return dc.store.Delete(handle)
}

func (dc *Cache) Lookup(name string) (Handle, bool) {
	return dc.store.Lookup(name)
}
