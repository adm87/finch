package aseprite

import (
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"sync"

	"github.com/adm87/finch/pkg/logger"
	"github.com/adm87/finch/pkg/resources"
	"github.com/adm87/finch/pkg/storage"
	"github.com/adm87/finch/pkg/types/errs"
	"github.com/adm87/finch/pkg/types/slotmap"
)

const aseprite_library = "aseprite library"

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

// LibraryID uniquely identifies an Aseprite library.
type LibraryID uint64

// Aseprite manages Aseprite libraries and their atlases.
//
// Ownership is all-or-nothing: if a library's content was partially preloaded,
// ClearLibrary with UnloadModeIfOwned may unload the preloaded parts too.
// Preload a library's content entirely or not at all.
type Aseprite interface {
	// CreateLibrary returns the ID of the named library, loading and decoding it on first use.
	// If the library's content isn't loaded, it is loaded only when dynamicLoad is true.
	CreateLibrary(name string, mode LoadMode) (LibraryID, error)

	// ClearLibrary removes the named library.
	// If dynamicUnload is true, its content is unloaded when nothing else needs it.
	ClearLibrary(name string, mode UnloadMode) error

	// GetLibrary returns the library with the given ID.
	// The Atlas is shared and must not be modified.
	GetLibrary(id LibraryID) (LibraryInfo, error)
}

// LibraryInfo contains information about an Aseprite library, including its atlas, image handle, and data handle.
type LibraryInfo struct {
	Atlas       *Atlas
	ImageHandle uint64
	DataHandle  uint64
	owned       bool
}

type ase struct {
	res    resources.Resources
	logger logger.Logger

	dc        *storage.DataCache
	ic        *storage.ImageCache
	libraries *slotmap.SlotMap[LibraryInfo]
	config    *Config

	libsByName map[string]LibraryID

	mu sync.Mutex
}

// New creates a new Aseprite instance with the given resources, logger, data cache, image cache, and configuration.
func New(res resources.Resources, log logger.Logger, dc *storage.DataCache, ic *storage.ImageCache, config *Config) (Aseprite, error) {
	var errList []error
	if res == nil {
		errList = append(errList, errs.ArgumentNil("res"))
	}
	if dc == nil {
		errList = append(errList, errs.ArgumentNil("dc"))
	}
	if ic == nil {
		errList = append(errList, errs.ArgumentNil("ic"))
	}
	if config == nil {
		errList = append(errList, errs.ArgumentNil("config"))
	} else {
		errList = append(errList, validateConfig(config))
	}
	if err := errors.Join(errList...); err != nil {
		return nil, err
	}
	if log == nil {
		log = logger.N
	}
	return &ase{
		res:        res,
		logger:     log.Named("aseprite"),
		dc:         dc,
		ic:         ic,
		config:     config,
		libraries:  slotmap.New[LibraryInfo](128),
		libsByName: make(map[string]LibraryID),
	}, nil
}

func validateConfig(cfg *Config) error {
	var errList []error
	for name, lib := range cfg.Libraries {
		if err := lib.Validate(); err != nil {
			errList = append(errList, errs.Argument("config", fmt.Sprintf("library %q", name)).Wrap(err))
		}
	}
	return errors.Join(errList...)
}

func (a *ase) CreateLibrary(name string, mode LoadMode) (LibraryID, error) {
	a.mu.Lock()
	defer a.mu.Unlock()

	if id, exists := a.libsByName[name]; exists {
		return id, nil
	}

	library, err := a.getLibraryConfig(name)
	if err != nil {
		return 0, err
	}

	loaded, err := a.isLibraryLoaded(library)
	if err != nil {
		return 0, fmt.Errorf("check library %q: %w", name, err)
	}

	owned := false
	if !loaded {
		if mode == LoadModeRequired {
			return 0, errs.InvalidOperation(fmt.Sprintf("library %q is not loaded and mode is LoadModeRequired", name))
		}
		if err := a.loadLibrary(library); err != nil && !errs.IsAlreadyExists(err) {
			return 0, fmt.Errorf("load library %q: %w", name, err)
		}
		owned = true
	}

	info, err := a.decodeLibrary(library)
	if err != nil {
		if owned {
			err = errors.Join(err, a.unloadLibrary(library))
		}
		return 0, fmt.Errorf("library %q: %w", name, err)
	}
	info.owned = owned

	id := LibraryID(a.libraries.Insert(info).Pack())
	a.libsByName[name] = id
	return id, nil
}

func (a *ase) ClearLibrary(name string, mode UnloadMode) error {
	a.mu.Lock()
	defer a.mu.Unlock()

	id, exists := a.libsByName[name]
	if !exists {
		return errs.NotFound("aseprite library", name)
	}

	info, err := a.libraries.Delete(slotmap.Unpack(uint64(id)))
	delete(a.libsByName, name)
	if err != nil {
		return fmt.Errorf("clear library %q: %w", name, err)
	}

	if mode != UnloadModeIfOwned || !info.owned {
		return nil
	}

	cfg, err := a.getLibraryConfig(name)
	if err != nil {
		return err
	}
	if err := a.unloadLibrary(cfg); err != nil {
		return fmt.Errorf("unload library %q: %w", name, err)
	}
	return nil
}

func (a *ase) GetLibrary(id LibraryID) (LibraryInfo, error) {
	a.mu.Lock()
	defer a.mu.Unlock()

	info, err := a.libraries.Get(slotmap.Unpack(uint64(id)))
	if err != nil {
		return LibraryInfo{}, errs.NotFound(aseprite_library, strconv.FormatUint(uint64(id), 10)).Wrap(err)
	}
	return info, nil
}

func (a *ase) getLibraryConfig(name string) (*LibraryConfig, error) {
	if name == "" {
		return nil, errs.Argument("name", "library name cannot be empty")
	}
	library, exists := a.config.Libraries[name]
	if !exists {
		return nil, errs.NotFound(aseprite_library, name)
	}
	return &library, nil
}

func (a *ase) decodeLibrary(library *LibraryConfig) (LibraryInfo, error) {
	atlasPath := library.Atlas.String()

	dataHandle, err := a.res.GetHandle(library.Atlas)
	if err != nil {
		return LibraryInfo{}, fmt.Errorf("atlas %q: %w", atlasPath, err)
	}
	imageHandle, err := a.res.GetHandle(library.Image)
	if err != nil {
		return LibraryInfo{}, fmt.Errorf("image %q: %w", library.Image.String(), err)
	}
	raw, err := a.dc.Get(dataHandle)
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
		a.logger.Warn("dropping missing tags", logger.String("atlas", atlasPath), logger.Strings("tags", dropped))
	}
	if err := info.Atlas.Validate(); err != nil {
		return LibraryInfo{}, fmt.Errorf("atlas %q: %w", atlasPath, err)
	}
	return info, nil
}

func (a *ase) isLibraryLoaded(cfg *LibraryConfig) (bool, error) {
	for _, r := range cfg.Resources() {
		_, err := a.res.GetHandle(r)
		if errs.IsNotFound(err) {
			return false, nil
		}
		if err != nil {
			return false, err
		}
	}
	return true, nil
}

func (a *ase) loadLibrary(cfg *LibraryConfig) error {
	bundle := resources.ResourceBundle{
		FileSystem: resources.FileSystem(cfg.FileSystem),
		Resources:  cfg.Resources(),
	}
	return bundle.Load(a.res)
}

func (a *ase) unloadLibrary(cfg *LibraryConfig) error {
	bundle := resources.ResourceBundle{
		FileSystem: resources.FileSystem(cfg.FileSystem),
		Resources:  cfg.Resources(),
	}
	return bundle.Unload(a.res)
}
