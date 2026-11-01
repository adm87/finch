package aseprite

import (
	"encoding/json"
	"fmt"
	"io/fs"

	"github.com/adm87/finch/pkg/errs"
	"github.com/adm87/finch/pkg/images"
	"github.com/adm87/finch/pkg/logger"
	"github.com/adm87/finch/pkg/resource"
)

const libraryKind = "aseprite library"

// Handle uniquely identifies an Aseprite library.
type Handle uint64

// LibraryInfo is a decoded library and the handle of its image.
type LibraryInfo struct {
	Atlas       *Atlas
	ImageHandle images.Handle
}

// Load builds the named library from src, or returns it if it's already
// loaded from there.
func (p *Plugin) Load(src resource.Source, name string) (Handle, error) {
	key := resource.Key(src, name)
	if handle, exists := p.libraries.Lookup(key); exists {
		return handle, nil
	}

	library, err := p.getLibraryConfig(name)
	if err != nil {
		return 0, err
	}
	fsys, err := p.sources.FS(src)
	if err != nil {
		return 0, fmt.Errorf("library %q: %w", name, err)
	}

	atlas, err := p.decodeAtlas(fsys, library.Atlas)
	if err != nil {
		return 0, fmt.Errorf("library %q: %w", name, err)
	}

	imageHandle, err := p.images.Load(resource.Key(src, library.Image), fsys, library.Image)
	if err != nil {
		return 0, fmt.Errorf("library %q: %w", name, err)
	}
	if err := p.images.SliceFrames(imageHandle, atlas.Frames.FrameRects()); err != nil {
		return 0, fmt.Errorf("slice image %q: %w", library.Image, err)
	}

	return p.libraries.Insert(key, LibraryInfo{Atlas: atlas, ImageHandle: imageHandle})
}

// Unload forgets the library loaded from src. Its image stays in the image cache.
func (p *Plugin) Unload(src resource.Source, name string) error {
	handle, exists := p.libraries.Lookup(resource.Key(src, name))
	if !exists {
		return errs.NotFound(libraryKind, name)
	}
	if _, err := p.libraries.Delete(handle); err != nil {
		return fmt.Errorf("unload library %q: %w", name, err)
	}
	return nil
}

func (p *Plugin) Library(handle Handle) (LibraryInfo, error) {
	return p.libraries.Get(handle)
}

func (p *Plugin) getLibraryConfig(name string) (*LibraryConfig, error) {
	if name == "" {
		return nil, errs.Argument("name", "cannot be empty")
	}
	library, exists := p.config.Libraries[name]
	if !exists {
		return nil, errs.NotFound(libraryKind, name)
	}
	return &library, nil
}

func (p *Plugin) decodeAtlas(fsys fs.FS, name string) (*Atlas, error) {
	raw, err := fs.ReadFile(fsys, name)
	if err != nil {
		return nil, fmt.Errorf("read %q: %w", name, err)
	}

	var atlas *Atlas
	if err := json.Unmarshal(raw, &atlas); err != nil {
		return nil, errs.Format(name, "invalid aseprite json").Wrap(err)
	}
	if atlas == nil {
		return nil, errs.Format(name, "atlas is empty")
	}
	if dropped := atlas.DropMissingTags(); len(dropped) > 0 {
		p.logger.Warn("dropping missing tags", logger.String("atlas", name), logger.Strings("tags", dropped))
	}
	if err := atlas.Validate(); err != nil {
		return nil, fmt.Errorf("atlas %q: %w", name, err)
	}

	atlas.fitTagsToFrames()
	return atlas, nil
}
