package assets

import (
	"errors"
	"fmt"
	"io"
	"io/fs"
	"slices"
	"sync"

	"github.com/adm87/finch/pkg/types/errs"
)

type FileSystem string

type Importer interface {
	Import(reader io.Reader) (uint64, error)
	Delete(handle uint64) error
}

type Assets struct {
	importers   map[string]Importer
	filesystems map[FileSystem]fs.FS
	handles     map[Asset]uint64
	mu          sync.Mutex
}

func NewAssets() *Assets {
	return &Assets{
		importers:   make(map[string]Importer),
		filesystems: make(map[FileSystem]fs.FS),
		handles:     make(map[Asset]uint64),
	}
}

// GetHandle retrieves the handle for the given asset, returning an error if it is not loaded.
func (a *Assets) GetHandle(asset Asset) (uint64, error) {
	a.mu.Lock()
	defer a.mu.Unlock()

	handle, exists := a.handles[asset]
	if !exists {
		return 0, errs.NotFound("asset", asset.String())
	}

	return handle, nil
}

func (a *Assets) GetFileSystem(name FileSystem) (fs.FS, error) {
	a.mu.Lock()
	defer a.mu.Unlock()

	fsys, exists := a.filesystems[name]
	if !exists {
		return nil, errs.NotFound("file system", string(name))
	}

	return fsys, nil
}

func (a *Assets) RegisterImporter(importer Importer, assetTypes []string) error {
	if importer == nil {
		return errs.ArgumentNil("importer")
	}

	types := make([]string, 0, len(assetTypes))
	for _, assetType := range assetTypes {
		if normalized := normalizeType(assetType); normalized != "" {
			types = append(types, normalized)
		}
	}

	a.mu.Lock()
	defer a.mu.Unlock()

	for i, normalized := range types {
		if _, exists := a.importers[normalized]; exists || slices.Contains(types[:i], normalized) {
			return errs.AlreadyExists("importer", normalized)
		}
	}

	for _, normalized := range types {
		a.importers[normalized] = importer
	}

	return nil
}

func (a *Assets) RegisterFileSystem(name FileSystem, filesystem fs.FS) error {
	if filesystem == nil {
		return errs.ArgumentNil("filesystem")
	}

	a.mu.Lock()
	defer a.mu.Unlock()

	if _, exists := a.filesystems[name]; exists {
		return errs.AlreadyExists("file system", string(name))
	}

	a.filesystems[name] = filesystem
	return nil
}

func (a *Assets) Load(filesystem fs.FS, batch ...Asset) error {
	if filesystem == nil {
		return errs.ArgumentNil("filesystem")
	}

	return a.run(batch, func(asset Asset) error {
		return a.load(filesystem, asset)
	})
}

func (a *Assets) Unload(batch ...Asset) error {
	return a.run(batch, func(asset Asset) error {
		return a.unload(asset)
	})
}

func (a *Assets) run(batch []Asset, operation func(Asset) error) error {
	requests := a.distinct(batch)
	if len(requests) == 0 {
		return nil
	}

	var errs []error
	for _, asset := range requests {
		if err := operation(asset); err != nil {
			errs = append(errs, err)
		}
	}

	return errors.Join(errs...)
}

func (a *Assets) load(fsys fs.FS, asset Asset) error {
	name := asset.String()

	a.mu.Lock()
	_, loaded := a.handles[asset]
	importer, ok := a.importers[asset.Type()]
	a.mu.Unlock()

	if loaded {
		return errs.AlreadyExists("asset", name)
	}
	if !ok {
		return errs.NotFound("importer", asset.Type())
	}

	file, err := fsys.Open(name)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return errs.NotFound("asset", name).Wrap(err)
		}
		return fmt.Errorf("open %q: %w", name, err)
	}
	defer file.Close()

	handle, err := importer.Import(file)
	if err != nil {
		return fmt.Errorf("import %q: %w", name, err)
	}

	a.mu.Lock()
	_, loaded = a.handles[asset]
	if !loaded {
		a.handles[asset] = handle
	}
	a.mu.Unlock()

	if !loaded {
		return nil
	}

	dup := errs.AlreadyExists("asset", name)
	if err := importer.Delete(handle); err != nil {
		// The asset was loaded concurrently by another goroutine.
		// Discard the handle we just imported.
		return errors.Join(dup, fmt.Errorf("discard duplicate %q: %w", name, err))
	}
	return dup
}

func (a *Assets) unload(asset Asset) error {
	a.mu.Lock()
	importer, hasImporter := a.importers[asset.Type()]
	handle, loaded := a.handles[asset]
	if loaded && hasImporter {
		delete(a.handles, asset)
	}
	a.mu.Unlock()

	if !loaded {
		return errs.NotFound("asset", asset.String())
	}
	if !hasImporter {
		return errs.NotFound("importer", asset.Type())
	}

	if err := importer.Delete(handle); err != nil {
		return fmt.Errorf("unload %q: %w", asset, err)
	}
	return nil
}

func (a *Assets) distinct(batch []Asset) []Asset {
	seen := make(map[Asset]struct{}, len(batch))
	distinct := make([]Asset, 0, len(batch))
	for _, asset := range batch {
		if _, exists := seen[asset]; !exists {
			seen[asset] = struct{}{}
			distinct = append(distinct, asset)
		}
	}
	return distinct
}
