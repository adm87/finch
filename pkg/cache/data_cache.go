package cache

import (
	"github.com/adm87/finch/pkg/types/slotmap"
)

type DataCache struct {
	store *slotmap.SlotMap[[]byte]
}

func NewDataCache() *DataCache {
	return &DataCache{
		store: slotmap.New[[]byte](256),
	}
}

func (dc *DataCache) Get(handle uint64) ([]byte, error) {
	img, err := dc.store.Get(slotmap.Unpack(handle))
	if err != nil {
		return nil, cacheError("get", "data", handle, err)
	}
	return img, nil
}

func (dc *DataCache) Set(handle uint64, data []byte) error {
	if err := dc.store.Set(slotmap.Unpack(handle), data); err != nil {
		return cacheError("set", "data", handle, err)
	}
	return nil
}

func (dc *DataCache) Put(raw []byte) uint64 {
	k := dc.store.Insert(raw)
	return k.Pack()
}

func (dc *DataCache) Delete(handle uint64) ([]byte, error) {
	data, err := dc.store.Delete(slotmap.Unpack(handle))
	if err != nil {
		return nil, cacheError("delete", "data", handle, err)
	}
	return data, nil
}
