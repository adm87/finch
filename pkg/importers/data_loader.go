package importers

import "io"

var dataTypes = []string{"json", "yml", "yaml"}

func DataTypes() []string {
	return dataTypes
}

type DataStorage interface {
	Put(raw []byte) uint64
	Delete(handle uint64) ([]byte, error)
}

type DataLoader struct {
	store DataStorage
}

func NewDataLoader(store DataStorage) *DataLoader {
	return &DataLoader{
		store: store,
	}
}

func (dl *DataLoader) Import(reader io.Reader) (uint64, error) {
	raw, err := io.ReadAll(reader)
	if err != nil {
		return 0, err
	}
	return dl.store.Put(raw), nil
}

func (dl *DataLoader) Delete(handle uint64) error {
	_, err := dl.store.Delete(handle)
	return err
}
