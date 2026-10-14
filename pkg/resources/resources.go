package resources

import (
	"errors"
	"fmt"
	"io/fs"
	"slices"
	"sync"

	"github.com/adm87/finch/pkg/types/slotmap"
	"golang.org/x/sync/errgroup"
)

var (
	ErrNotFound                  = errors.New("resource not found")
	ErrInvalidFileSystemName     = errors.New("invalid file system name")
	ErrImporterNotFound          = errors.New("importer not found")
	ErrNilImporter               = errors.New("nil importer")
	ErrNilFileSystem             = errors.New("nil file system")
	ErrDuplicateImporterRegister = errors.New("duplicate importer register")
	ErrDuplicateFileSystem       = errors.New("duplicate file system")
	ErrAlreadyLoaded             = errors.New("resource already loaded")
	ErrReadFile                  = errors.New("failed to read file")
	ErrImportFailed              = errors.New("failed to import resource")
	ErrNotLoaded                 = errors.New("resource not loaded")
	ErrUnloadFailed              = errors.New("failed to unload resource")
)

func ConvertToResourceError(err error) error {
	switch {
	case errors.Is(err, slotmap.ErrNotFound):
		return ErrNotFound
	}
	return err
}

type Resources interface {
	Load(filesystem fs.FS, batch ...Resource) error

	Unload(batch ...Resource) error

	RegisterImporter(importer ResourceImporter, resourceTypes []string) error

	AddFileSystem(name string, filesystem fs.FS) error

	GetHandle(resource Resource) (uint64, error)

	GetFileSystem(name string) (fs.FS, error)
}

type res struct {
	importers   map[string]ResourceImporter
	filesystems map[string]fs.FS
	handles     map[Resource]uint64
	concurrency int
	mu          sync.Mutex
}

func NewResources() Resources {
	return &res{
		importers:   make(map[string]ResourceImporter),
		filesystems: make(map[string]fs.FS),
		handles:     make(map[Resource]uint64),
		concurrency: 100, // TODO: Make this configurable
	}
}

// GetHandle retrieves the handle for the given resource, returning an error if it is not loaded.
func (r *res) GetHandle(resource Resource) (uint64, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	handle, exists := r.handles[resource]
	if !exists {
		return 0, fmt.Errorf("%w: %s", ErrNotFound, resource.String())
	}

	return handle, nil
}

func (r *res) GetFileSystem(name string) (fs.FS, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	fsys, exists := r.filesystems[name]
	if !exists {
		return nil, fmt.Errorf("%w: %s", ErrInvalidFileSystemName, name)
	}

	return fsys, nil
}

func (r *res) RegisterImporter(importer ResourceImporter, resourceTypes []string) error {
	if importer == nil {
		return ErrNilImporter
	}

	types := make([]string, 0, len(resourceTypes))
	for _, resourceType := range resourceTypes {
		if normalized := normalizeType(resourceType); normalized != "" {
			types = append(types, normalized)
		}
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	for i, normalized := range types {
		if _, exists := r.importers[normalized]; exists || slices.Contains(types[:i], normalized) {
			return fmt.Errorf("%w: %s", ErrDuplicateImporterRegister, normalized)
		}
	}

	for _, normalized := range types {
		r.importers[normalized] = importer
	}

	return nil
}

func (r *res) AddFileSystem(name string, filesystem fs.FS) error {
	if filesystem == nil {
		return ErrNilFileSystem
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.filesystems[name]; exists {
		return fmt.Errorf("%w: %s", ErrDuplicateFileSystem, name)
	}

	r.filesystems[name] = filesystem
	return nil
}

func (r *res) Load(filesystem fs.FS, batch ...Resource) error {
	if filesystem == nil {
		return ErrNilFileSystem
	}

	return r.run(batch, func(resource Resource) error {
		return r.loadResource(filesystem, resource)
	})
}

func (r *res) Unload(batch ...Resource) error {
	return r.run(batch, func(resource Resource) error {
		return r.unloadResource(resource)
	})
}

func (r *res) run(batch []Resource, operation func(Resource) error) error {
	requests := r.distinct(batch)
	if len(requests) == 0 {
		return nil
	}

	var errs []error
	var mu sync.Mutex
	var grp errgroup.Group

	grp.SetLimit(r.concurrency)
	for _, resource := range requests {
		grp.Go(func() error {
			if err := operation(resource); err != nil {
				mu.Lock()
				errs = append(errs, err)
				mu.Unlock()
			}
			return nil
		})
	}
	_ = grp.Wait()

	if len(errs) > 0 {
		return fmt.Errorf("resource errors: %w", errors.Join(errs...))
	}
	return nil
}

func (r *res) loadResource(fsys fs.FS, resource Resource) error {
	r.mu.Lock()
	_, loaded := r.handles[resource]
	importer, ok := r.importers[resource.Type()]
	r.mu.Unlock()

	if loaded {
		return fmt.Errorf("%w: %s", ErrAlreadyLoaded, resource)
	}
	if !ok {
		return fmt.Errorf("%w: %s", ErrImporterNotFound, resource.Type())
	}

	data, err := fs.ReadFile(fsys, resource.String())
	if err != nil {
		return fmt.Errorf("%w: %s: %w", ErrReadFile, resource, err)
	}

	handle, err := importer.Import(data)
	if err != nil {
		return fmt.Errorf("%w: %s: %w", ErrImportFailed, resource, err)
	}

	r.mu.Lock()
	_, loaded = r.handles[resource]
	if !loaded {
		r.handles[resource] = handle
	}
	r.mu.Unlock()

	if loaded {
		// The resource was loaded concurrently by another goroutine. Discard the handle we just imported.
		if err := importer.Delete(handle); err != nil {
			return fmt.Errorf("discard duplicate %s: %w", resource, err)
		}
		return fmt.Errorf("%w: %s", ErrAlreadyLoaded, resource)
	}
	return nil
}

func (r *res) unloadResource(resource Resource) error {
	r.mu.Lock()
	importer, hasImporter := r.importers[resource.Type()]
	handle, loaded := r.handles[resource]
	if loaded && hasImporter {
		delete(r.handles, resource)
	}
	r.mu.Unlock()

	if !loaded {
		return fmt.Errorf("%w: %s", ErrNotLoaded, resource)
	}
	if !hasImporter {
		return fmt.Errorf("%w: %s", ErrImporterNotFound, resource.Type())
	}

	if err := importer.Delete(handle); err != nil {
		return fmt.Errorf("%w: %s: %w", ErrUnloadFailed, resource, err)
	}
	return nil
}

func (r *res) distinct(batch []Resource) []Resource {
	seen := make(map[Resource]struct{}, len(batch))
	distinct := make([]Resource, 0, len(batch))
	for _, resource := range batch {
		if _, exists := seen[resource]; !exists {
			seen[resource] = struct{}{}
			distinct = append(distinct, resource)
		}
	}
	return distinct
}
