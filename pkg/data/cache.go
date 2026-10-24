package data

import (
	"github.com/adm87/finch/pkg/slotmap"
)

// Handle refers to data in a Cache.
type Handle uint64

type Cache struct {
	store *slotmap.SlotMap[[]byte]
}

func NewCache() *Cache {
	return &Cache{
		store: slotmap.New[[]byte](256),
	}
}

func (dc *Cache) Get(handle Handle) ([]byte, error) {
	data, err := dc.store.Get(slotmap.Unpack(uint64(handle)))
	if err != nil {
		return nil, slotmap.HandleError("get", "data", uint64(handle), err)
	}
	return data, nil
}

func (dc *Cache) Set(handle Handle, data []byte) error {
	if err := dc.store.Set(slotmap.Unpack(uint64(handle)), data); err != nil {
		return slotmap.HandleError("set", "data", uint64(handle), err)
	}
	return nil
}

func (dc *Cache) Put(raw []byte) Handle {
	k := dc.store.Insert(raw)
	return Handle(k.Pack())
}

func (dc *Cache) Delete(handle Handle) ([]byte, error) {
	data, err := dc.store.Delete(slotmap.Unpack(uint64(handle)))
	if err != nil {
		return nil, slotmap.HandleError("delete", "data", uint64(handle), err)
	}
	return data, nil
}
