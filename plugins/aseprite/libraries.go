package aseprite

import (
	"encoding/json"
	"errors"
	"fmt"
	"strconv"

	"github.com/adm87/finch/pkg/assets"
	"github.com/adm87/finch/pkg/data"
	"github.com/adm87/finch/pkg/errs"
	"github.com/adm87/finch/pkg/images"
	"github.com/adm87/finch/pkg/logger"
	"github.com/adm87/finch/pkg/slotmap"
)

const libraryKind = "aseprite library"

// LoadMode controls whether CreateLibrary may load missing content.
type LoadMode int

const (
	// LoadModeRequired fails if the library's content isn't already loaded.
	LoadModeRequired LoadMode = iota
	// LoadModeDynamic loads the library's content if it isn't loaded.
	LoadModeDynamic
)

// UnloadMode controls whether ClearLibrary unloads the library's content.
type UnloadMode int

const (
	// UnloadModeKeep leaves the content loaded.
	UnloadModeKeep UnloadMode = iota
	// UnloadModeIfOwned unloads the content only if CreateLibrary loaded it.
	UnloadModeIfOwned
)

// Handle uniquely identifies an Aseprite library.
type Handle uint64

// LibraryInfo is a decoded library and the handles of its content.
type LibraryInfo struct {
	Atlas       *Atlas
	ImageHandle images.Handle
	DataHandle  data.Handle
	owned       bool
}

type Libraries struct {
	logger     logger.Logger
	as         *assets.Assets
	dc         *data.Cache
	ic         *images.Cache
	libraries  *slotmap.SlotMap[LibraryInfo]
	config     *Config
	libsByName map[string]Handle
}

func NewLibraries(as *assets.Assets, log logger.Logger, dc *data.Cache, ic *images.Cache, config *Config) *Libraries {
	return &Libraries{
		logger:     log.Named("aseprite"),
		as:         as,
		dc:         dc,
		ic:         ic,
		config:     config,
		libraries:  slotmap.New[LibraryInfo](128),
		libsByName: make(map[string]Handle),
	}
}

func (l *Libraries) CreateLibrary(name string, mode LoadMode) (Handle, error) {
	if id, exists := l.libsByName[name]; exists {
		return id, nil
	}

	library, err := l.getLibraryConfig(name)
	if err != nil {
		return 0, err
	}

	loaded, err := l.isLibraryLoaded(library)
	if err != nil && !errs.IsNotFound(err) {
		return 0, fmt.Errorf("check library %q: %w", name, err)
	}

	owned := false
	if !loaded {
		if mode == LoadModeRequired {
			return 0, errs.InvalidOperation(fmt.Sprintf("library %q is not loaded and mode is LoadModeRequired", name))
		}
		if err := l.loadLibrary(library); err != nil && !errs.IsAlreadyExists(err) {
			return 0, fmt.Errorf("load library %q: %w", name, err)
		}
		owned = true
	}

	info, err := l.decodeLibrary(library)
	if err != nil {
		if owned {
			err = errors.Join(err, l.unloadLibrary(library))
		}
		return 0, fmt.Errorf("library %q: %w", name, err)
	}
	info.owned = owned

	if err := l.ic.SliceFrames(info.ImageHandle, info.Atlas.Frames.FrameRects()); err != nil {
		return 0, fmt.Errorf("slice image %q: %w", library.Image.String(), err)
	}

	id := Handle(l.libraries.Insert(info).Pack())
	l.libsByName[name] = id

	return id, nil
}

func (l *Libraries) ClearLibrary(name string, mode UnloadMode) error {
	id, exists := l.libsByName[name]
	if !exists {
		return errs.NotFound(libraryKind, name)
	}

	info, err := l.libraries.Delete(slotmap.Unpack(uint64(id)))
	delete(l.libsByName, name)
	if err != nil {
		return fmt.Errorf("clear library %q: %w", name, err)
	}

	if mode != UnloadModeIfOwned || !info.owned {
		return nil
	}

	cfg, err := l.getLibraryConfig(name)
	if err != nil {
		return err
	}
	if err := l.unloadLibrary(cfg); err != nil {
		return fmt.Errorf("unload library %q: %w", name, err)
	}
	return nil
}

func (l *Libraries) Library(id Handle) (LibraryInfo, error) {
	info, err := l.libraries.Get(slotmap.Unpack(uint64(id)))
	if err != nil {
		return LibraryInfo{}, errs.NotFound(libraryKind, strconv.FormatUint(uint64(id), 10)).Wrap(err)
	}
	return info, nil
}

func (l *Libraries) getLibraryConfig(name string) (*LibraryConfig, error) {
	if name == "" {
		return nil, errs.Argument("name", "cannot be empty")
	}
	library, exists := l.config.Libraries[name]
	if !exists {
		return nil, errs.NotFound(libraryKind, name)
	}
	return &library, nil
}

func (l *Libraries) decodeLibrary(library *LibraryConfig) (LibraryInfo, error) {
	atlasPath := library.Atlas.String()

	atlasAsset, err := l.as.Handle(library.Atlas)
	if err != nil {
		return LibraryInfo{}, fmt.Errorf("atlas %q: %w", atlasPath, err)
	}
	imageAsset, err := l.as.Handle(library.Image)
	if err != nil {
		return LibraryInfo{}, fmt.Errorf("image %q: %w", library.Image.String(), err)
	}
	dataHandle := data.Handle(atlasAsset)
	imageHandle := images.Handle(imageAsset)

	raw, err := l.dc.Get(dataHandle)
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
		l.logger.Warn("dropping missing tags", logger.String("atlas", atlasPath), logger.Strings("tags", dropped))
	}
	if err := info.Atlas.Validate(); err != nil {
		return LibraryInfo{}, fmt.Errorf("atlas %q: %w", atlasPath, err)
	}
	info.Atlas.fitTagsToFrames()
	return info, nil
}

func (l *Libraries) isLibraryLoaded(cfg *LibraryConfig) (bool, error) {
	for _, r := range cfg.Assets() {
		if _, err := l.as.Handle(r); err != nil {
			return false, err
		}
	}
	return true, nil
}

func (l *Libraries) loadLibrary(cfg *LibraryConfig) error {
	bundle := assets.AssetBundle{
		FileSystem: assets.FileSystem(cfg.FileSystem),
		Assets:     cfg.Assets(),
	}
	return bundle.Load(l.as)
}

func (l *Libraries) unloadLibrary(cfg *LibraryConfig) error {
	bundle := assets.AssetBundle{
		FileSystem: assets.FileSystem(cfg.FileSystem),
		Assets:     cfg.Assets(),
	}
	return bundle.Unload(l.as)
}
