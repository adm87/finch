package storage

var dataTypes = []string{"json", "yml", "yaml"}

func DataTypes() []string {
	return dataTypes
}

type DataLoader struct {
	cache *DataCache
}

func NewDataLoader(cache *DataCache) *DataLoader {
	return &DataLoader{
		cache: cache,
	}
}

func (dl *DataLoader) Import(raw []byte) (uint64, error) {
	return dl.cache.Put(raw), nil
}

func (dl *DataLoader) Delete(handle uint64) error {
	_, err := dl.cache.Delete(handle)
	return err
}
