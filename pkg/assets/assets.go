package assets

import (
	"errors"
	"fmt"
	"io"
	"io/fs"
	"slices"
	"sync"

	"github.com/adm87/finch/pkg/errs"
)

type FileSystem string

// Handle is an importer's handle for a loaded asset.
type Handle uint64

type Importer interface {
	Import(reader io.Reader) (Handle, error)
	Delete(handle Handle) error
}

type Assets struct {
	importers   map[string]Importer
	filesystems map[FileSystem]fs.FS
	handles     map[Asset]Handle
	mu          sync.Mutex
}

func NewAssets() *Assets {
	return &Assets{
		importers:   make(map[string]Importer),
		filesystems: make(map[FileSystem]fs.FS),
		handles:     make(map[Asset]Handle),
	}
}

// Handle returns a NotFoundError if the asset isn't loaded.
func (a *Assets) Handle(asset Asset) (Handle, error) {
	a.mu.Lock()
	defer a.mu.Unlock()

	handle, exists := a.handles[asset]
	if !exists {
		return 0, errs.NotFound("asset", asset.String())
	}

	return handle, nil
}

func (a *Assets) FileSystem(name FileSystem) (fs.FS, error) {
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

	var errList []error
	for _, asset := range requests {
		if err := operation(asset); err != nil {
			errList = append(errList, err)
		}
	}

	return errors.Join(errList...)
}

func (a *Assets) load(fsys fs.FS, asset Asset) error {
	name := asset.String()

	a.mu.Lock()
	defer a.mu.Unlock()

	if _, loaded := a.handles[asset]; loaded {
		return errs.AlreadyExists("asset", name)
	}
	importer, ok := a.importers[asset.Type()]
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
	defer func() { _ = file.Close() }()

	handle, err := importer.Import(file)
	if err != nil {
		return fmt.Errorf("import %q: %w", name, err)
	}
	a.handles[asset] = handle
	return nil
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
