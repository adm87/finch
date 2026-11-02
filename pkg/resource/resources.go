package resource

import (
	"fmt"
	"io/fs"

	"github.com/adm87/finch/pkg/errs"
)

// Source names a file system that resources are read from.
type Source string

// Key names a file by its source and path, so the same path in two sources
// stays two resources.
func Key(src Source, name string) string {
	return string(src) + ":" + name
}

// Resources reads files from named sources, so loaders don't care where
// a file comes from.
type Resources struct {
	sources map[Source]fs.FS
}

func New() *Resources {
	return &Resources{
		sources: make(map[Source]fs.FS),
	}
}

func (r *Resources) Mount(name Source, source fs.FS) error {
	if _, exists := r.sources[name]; exists {
		return errs.AlreadyExists("resource source", string(name))
	}

	r.sources[name] = source
	return nil
}

func (r *Resources) ReadFile(src Source, name string) ([]byte, error) {
	fsys, ok := r.sources[src]
	if !ok {
		return nil, errs.NotFound("resource source", string(src))
	}
	raw, err := fs.ReadFile(fsys, name)
	if err != nil {
		return nil, fmt.Errorf("read %s:%s: %w", src, name, err)
	}
	return raw, nil
}
