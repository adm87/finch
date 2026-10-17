package resources

import (
	"errors"
	"fmt"
	"io/fs"
	"slices"
	"sync"

	"github.com/adm87/finch/pkg/types/errs"
	"golang.org/x/sync/errgroup"
)

type FileSystem string

type Resources interface {
	Load(filesystem fs.FS, batch ...Resource) error

	Unload(batch ...Resource) error

	RegisterImporter(importer ResourceImporter, resourceTypes []string) error

	RegisterFileSystem(name FileSystem, filesystem fs.FS) error

	GetHandle(resource Resource) (uint64, error)

	GetFileSystem(name FileSystem) (fs.FS, error)
}

type res struct {
	importers   map[string]ResourceImporter
	filesystems map[FileSystem]fs.FS
	handles     map[Resource]uint64
	concurrency int
	mu          sync.Mutex
}

func NewResources() Resources {
	return &res{
		importers:   make(map[string]ResourceImporter),
		filesystems: make(map[FileSystem]fs.FS),
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
		return 0, errs.NotFound("resource", resource.String())
	}

	return handle, nil
}

func (r *res) GetFileSystem(name FileSystem) (fs.FS, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	fsys, exists := r.filesystems[name]
	if !exists {
		return nil, errs.NotFound("file system", string(name))
	}

	return fsys, nil
}

func (r *res) RegisterImporter(importer ResourceImporter, resourceTypes []string) error {
	if importer == nil {
		return errs.ArgumentNil("importer")
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
			return errs.AlreadyExists("importer", normalized)
		}
	}

	for _, normalized := range types {
		r.importers[normalized] = importer
	}

	return nil
}

func (r *res) RegisterFileSystem(name FileSystem, filesystem fs.FS) error {
	if filesystem == nil {
		return errs.ArgumentNil("filesystem")
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.filesystems[name]; exists {
		return errs.AlreadyExists("file system", string(name))
	}

	r.filesystems[name] = filesystem
	return nil
}

func (r *res) Load(filesystem fs.FS, batch ...Resource) error {
	if filesystem == nil {
		return errs.ArgumentNil("filesystem")
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
		return errors.Join(errs...)
	}
	return nil
}

func (r *res) loadResource(fsys fs.FS, resource Resource) error {
	name := resource.String()

	r.mu.Lock()
	_, loaded := r.handles[resource]
	importer, ok := r.importers[resource.Type()]
	r.mu.Unlock()

	if loaded {
		return errs.AlreadyExists("resource", name)
	}
	if !ok {
		return errs.NotFound("importer", resource.Type())
	}

	data, err := fs.ReadFile(fsys, name)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return errs.NotFound("resource", name).Wrap(err)
		}
		return fmt.Errorf("read %q: %w", name, err)
	}

	handle, err := importer.Import(data)
	if err != nil {
		return fmt.Errorf("import %q: %w", name, err)
	}

	r.mu.Lock()
	_, loaded = r.handles[resource]
	if !loaded {
		r.handles[resource] = handle
	}
	r.mu.Unlock()

	if !loaded {
		return nil
	}

	dup := errs.AlreadyExists("resource", name)
	if err := importer.Delete(handle); err != nil {
		// The resource was loaded concurrently by another goroutine.
		// Discard the handle we just imported.
		return errors.Join(dup, fmt.Errorf("discard duplicate %q: %w", name, err))
	}
	return dup
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
		return errs.NotFound("resource", resource.String())
	}
	if !hasImporter {
		return errs.NotFound("importer", resource.Type())
	}

	if err := importer.Delete(handle); err != nil {
		return fmt.Errorf("unload %q: %w", resource, err)
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
