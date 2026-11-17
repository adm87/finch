package assets

import (
	"errors"
	"fmt"
	"io/fs"
)

var ErrMountConflict = errors.New("assets: path already mounted")

// Content represents a collection of mounted filesystems.
type Content struct {
	files map[string]fs.FS
}

// Mount adds every file in fsys. If any path is already mounted, it adds none.
func (c *Content) Mount(fsys fs.FS) error {
	if fsys == nil {
		panic("assets: fsys is nil")
	}

	paths, err := mountPaths(fsys)
	if err != nil {
		return err
	}

	var conflicts []error
	for _, path := range paths {
		if _, ok := c.files[path]; ok {
			conflicts = append(conflicts, fmt.Errorf("mount %q: %w", path, ErrMountConflict))
		}
	}

	if len(conflicts) > 0 {
		return errors.Join(conflicts...)
	}

	if c.files == nil {
		c.files = make(map[string]fs.FS)
	}

	for _, path := range paths {
		c.files[path] = fsys
	}

	return nil
}

func (c *Content) Open(name string) (fs.File, error) {
	if !fs.ValidPath(name) {
		return nil, &fs.PathError{Op: "open", Path: name, Err: fs.ErrInvalid}
	}

	fsys, ok := c.files[name]
	if !ok {
		return nil, &fs.PathError{Op: "open", Path: name, Err: fs.ErrNotExist}
	}

	return fsys.Open(name)
}

func mountPaths(fsys fs.FS) ([]string, error) {
	var paths []string
	err := fs.WalkDir(fsys, ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !d.IsDir() {
			paths = append(paths, path)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return paths, nil
}
