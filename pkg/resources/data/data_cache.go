package data

import (
	"github.com/adm87/finch/pkg/resources"
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
	return img, resources.ConvertToResourceError(err)
}

func (dc *DataCache) Set(handle uint64, data []byte) error {
	err := dc.store.Set(slotmap.Unpack(handle), data)
	return resources.ConvertToResourceError(err)
}

func (dc *DataCache) Put(raw []byte) uint64 {
	k := dc.store.Insert(raw)
	return k.Pack()
}

func (dc *DataCache) Delete(handle uint64) ([]byte, error) {
	data, err := dc.store.Delete(slotmap.Unpack(handle))
	return data, resources.ConvertToResourceError(err)
}
