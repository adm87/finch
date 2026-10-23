package data

import (
	"github.com/adm87/finch/pkg/slotmap"
)

type Cache struct {
	store *slotmap.SlotMap[[]byte]
}

func NewCache() *Cache {
	return &Cache{
		store: slotmap.New[[]byte](256),
	}
}

func (dc *Cache) Get(handle uint64) ([]byte, error) {
	data, err := dc.store.Get(slotmap.Unpack(handle))
	if err != nil {
		return nil, slotmap.HandleError("get", "data", handle, err)
	}
	return data, nil
}

func (dc *Cache) Set(handle uint64, data []byte) error {
	if err := dc.store.Set(slotmap.Unpack(handle), data); err != nil {
		return slotmap.HandleError("set", "data", handle, err)
	}
	return nil
}

func (dc *Cache) Put(raw []byte) uint64 {
	k := dc.store.Insert(raw)
	return k.Pack()
}

func (dc *Cache) Delete(handle uint64) ([]byte, error) {
	data, err := dc.store.Delete(slotmap.Unpack(handle))
	if err != nil {
		return nil, slotmap.HandleError("delete", "data", handle, err)
	}
	return data, nil
}
