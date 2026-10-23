package data

import (
	"io"
	"slices"
)

var dataTypes = []string{"json", "yml", "yaml"}

func Types() []string {
	return slices.Clone(dataTypes)
}

type Storage interface {
	Put(raw []byte) uint64
	Delete(handle uint64) ([]byte, error)
}

type Importer struct {
	store Storage
}

func NewImporter(store Storage) *Importer {
	return &Importer{
		store: store,
	}
}

func (dl *Importer) Import(reader io.Reader) (uint64, error) {
	raw, err := io.ReadAll(reader)
	if err != nil {
		return 0, err
	}
	return dl.store.Put(raw), nil
}

func (dl *Importer) Delete(handle uint64) error {
	_, err := dl.store.Delete(handle)
	return err
}
