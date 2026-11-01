package resource

import (
	"io/fs"
	"maps"

	"github.com/adm87/finch/pkg/errs"
)

// Source names a file system that resources are read from.
type Source string

// Loader resolves sources to the file systems behind them.
type Loader struct {
	sources map[Source]fs.FS
}

func NewLoader(sources map[Source]fs.FS) *Loader {
	return &Loader{sources: maps.Clone(sources)}
}

// FS returns the file system src refers to.
func (l *Loader) FS(src Source) (fs.FS, error) {
	fsys, ok := l.sources[src]
	if !ok {
		return nil, errs.NotFound("source", string(src))
	}
	return fsys, nil
}

// Key names a file by its source and path, so the same path in two sources
// stays two resources.
func Key(src Source, name string) string {
	return string(src) + ":" + name
}
