package resources

type ResourceImporter interface {
	Import(data []byte) (uint64, error)
	Delete(handle uint64) error
}
