package aseprite

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/adm87/finch/pkg/assets"
	"github.com/adm87/finch/pkg/data"
	"github.com/adm87/finch/pkg/errs"
	"github.com/adm87/finch/pkg/images"
	"github.com/adm87/finch/pkg/logger"
)

const libraryKind = "aseprite library"

// Handle uniquely identifies an Aseprite library.
type Handle uint64

// LibraryInfo is a decoded library and the handles of its content.
type LibraryInfo struct {
	Atlas       *Atlas
	ImageHandle images.Handle
	DataHandle  data.Handle
	owned       bool
}

// Load builds the named library from the config, loading its content if it
// isn't loaded yet, or returns it if it's already loaded.
func (p *Plugin) Load(name string) (Handle, error) {
	if handle, exists := p.libraries.Lookup(name); exists {
		return handle, nil
	}

	library, err := p.getLibraryConfig(name)
	if err != nil {
		return 0, err
	}

	owned := false
	if !p.isLibraryLoaded(library) {
		if err := p.loadLibrary(library); err != nil && !errs.IsAlreadyExists(err) {
			return 0, fmt.Errorf("load library %q: %w", name, err)
		}
		owned = true
	}

	info, err := p.decodeLibrary(library)
	if err != nil {
		if owned {
			err = errors.Join(err, p.unloadLibrary(library))
		}
		return 0, fmt.Errorf("library %q: %w", name, err)
	}
	info.owned = owned

	if err := p.images.SliceFrames(info.ImageHandle, info.Atlas.Frames.FrameRects()); err != nil {
		return 0, fmt.Errorf("slice image %q: %w", library.Image.String(), err)
	}

	return p.libraries.Insert(name, info)
}

// Unload forgets the library and unloads its content if its Load loaded it.
func (p *Plugin) Unload(name string) error {
	handle, exists := p.libraries.Lookup(name)
	if !exists {
		return errs.NotFound(libraryKind, name)
	}

	info, err := p.libraries.Delete(handle)
	if err != nil {
		return fmt.Errorf("unload library %q: %w", name, err)
	}

	if !info.owned {
		return nil
	}

	cfg, err := p.getLibraryConfig(name)
	if err != nil {
		return err
	}
	if err := p.unloadLibrary(cfg); err != nil {
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

func (p *Plugin) decodeLibrary(library *LibraryConfig) (LibraryInfo, error) {
	atlasPath := library.Atlas.String()

	imageHandle, exists := p.images.Lookup(library.Image.String())
	if !exists {
		return LibraryInfo{}, errs.NotFound("image", library.Image.String())
	}

	dataHandle, exists := p.data.Lookup(atlasPath)
	if !exists {
		return LibraryInfo{}, errs.NotFound("data", atlasPath)
	}

	raw, err := p.data.Get(dataHandle)
	if err != nil {
		return LibraryInfo{}, fmt.Errorf("atlas %q: %w", atlasPath, err)
	}

	info := LibraryInfo{ImageHandle: imageHandle, DataHandle: dataHandle}
	if err := json.Unmarshal(raw, &info.Atlas); err != nil {
		return LibraryInfo{}, errs.Format(atlasPath, "invalid aseprite json").Wrap(err)
	}
	if info.Atlas == nil {
		return LibraryInfo{}, errs.Format(atlasPath, "atlas is empty")
	}
	if dropped := info.Atlas.DropMissingTags(); len(dropped) > 0 {
		p.logger.Warn("dropping missing tags", logger.String("atlas", atlasPath), logger.Strings("tags", dropped))
	}
	if err := info.Atlas.Validate(); err != nil {
		return LibraryInfo{}, fmt.Errorf("atlas %q: %w", atlasPath, err)
	}

	info.Atlas.fitTagsToFrames()
	return info, nil
}

func (p *Plugin) isLibraryLoaded(cfg *LibraryConfig) bool {
	for _, asset := range cfg.Assets() {
		if !p.assets.Loaded(asset) {
			return false
		}
	}
	return true
}

func (p *Plugin) loadLibrary(cfg *LibraryConfig) error {
	bundle := assets.AssetBundle{
		FileSystem: assets.FileSystem(cfg.FileSystem),
		Assets:     cfg.Assets(),
	}
	return bundle.Load(p.assets)
}

func (p *Plugin) unloadLibrary(cfg *LibraryConfig) error {
	bundle := assets.AssetBundle{
		FileSystem: assets.FileSystem(cfg.FileSystem),
		Assets:     cfg.Assets(),
	}
	return bundle.Unload(p.assets)
}
