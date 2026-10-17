package data

import (
	"errors"
	"fmt"
	"strconv"

	"github.com/adm87/finch/pkg/types/errs"
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
		return nil, storeError("get", handle, err)
	}
	return img, nil
}

func (dc *DataCache) Set(handle uint64, data []byte) error {
	if err := dc.store.Set(slotmap.Unpack(handle), data); err != nil {
		return storeError("set", handle, err)
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
		return nil, storeError("delete", handle, err)
	}
	return data, nil
}

func storeError(op string, handle uint64, err error) error {
	key := strconv.FormatUint(handle, 10)
	switch {
	case errors.Is(err, slotmap.ErrInvalidKey):
		return errs.Argument("handle", fmt.Sprintf("malformed image handle %s", key)).Wrap(err)
	case errors.Is(err, slotmap.ErrNotFound):
		return errs.NotFound("image", key).Wrap(err)
	default:
		return fmt.Errorf("%s image %s: %w", op, key, err)
	}
}
