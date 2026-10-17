package aseprite

import (
	"encoding/json"
	"errors"
	"fmt"
	"sync"

	"github.com/adm87/finch/pkg/images"
	"github.com/adm87/finch/pkg/logger"
	"github.com/adm87/finch/pkg/resources"
	"github.com/adm87/finch/pkg/resources/data"
	"github.com/adm87/finch/pkg/types/errs"
	"github.com/adm87/finch/pkg/types/slotmap"
)

type Aseprite interface {
	CreateLibrary(id LibraryID, dynamicLoad bool) (uint64, error)
	ClearLibrary(id LibraryID, dynamicUnload bool) error
}

type Atlases map[uint64]Atlas

type ase struct {
	res       resources.Resources
	logger    logger.Logger
	dc        *data.DataCache
	ic        *images.ImageCache
	config    *Config
	atlases   *slotmap.SlotMap[*Atlas]
	libraries map[LibraryID]uint64
	mu        sync.Mutex
}

func New(res resources.Resources, log logger.Logger, dc *data.DataCache, ic *images.ImageCache, config *Config) (Aseprite, error) {
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
	return &ase{
		res:       res,
		logger:    log.Named("aseprite"),
		dc:        dc,
		ic:        ic,
		config:    config,
		atlases:   slotmap.New[*Atlas](128),
		libraries: make(map[LibraryID]uint64),
	}, nil
}

func validateConfig(cfg *Config) error {
	var errList []error
	for id, lib := range cfg.Libraries {
		if lib.Manifest == "" {
			errList = append(errList, errs.Argument("config", fmt.Sprintf("library %q: manifest is missing", id)))
		}
		if lib.Content.FileSystem == "" {
			errList = append(errList, errs.Argument("config", fmt.Sprintf("library %q: content file system is missing", id)))
		}
		if len(lib.Content.Resources) == 0 {
			errList = append(errList, errs.Argument("config", fmt.Sprintf("library %q: content has no resources", id)))
		}
	}
	return errors.Join(errList...)
}

func (a *ase) CreateLibrary(id LibraryID, dynamicLoad bool) (uint64, error) {
	a.mu.Lock()
	defer a.mu.Unlock()

	if atlasID, exists := a.libraries[id]; exists {
		return atlasID, nil
	}

	library, err := a.GetLibrary(id)
	if err != nil {
		return 0, err
	}

	loaded, err := a.isLibraryLoaded(&library.Content)
	if err != nil {
		return 0, fmt.Errorf("check library %q: %w", id, err)
	}

	if !loaded {
		if !dynamicLoad {
			return 0, errs.InvalidOperation(fmt.Sprintf("library %q is not loaded and dynamic loading is disabled", id))
		}
		if err := library.Content.Load(a.res); err != nil && !errs.IsAlreadyExists(err) {
			return 0, fmt.Errorf("load library %q: %w", id, err)
		}
	}

	atlas, err := a.decodeAtlas(library.Manifest)
	if err != nil {
		return 0, fmt.Errorf("library %q: %w", id, err)
	}

	atlasID := a.atlases.Insert(atlas).Pack()
	a.libraries[id] = atlasID

	return atlasID, nil
}

func (a *ase) ClearLibrary(id LibraryID, dynamicUnload bool) error {
	return errs.NotImplemented("ClearLibrary")
}

func (a *ase) GetLibrary(id LibraryID) (*Library, error) {
	if id.IsEmpty() {
		return nil, errs.Argument("id", "library id cannot be empty")
	}
	library, exists := a.config.Libraries[id]
	if !exists {
		return nil, errs.NotFound("aseprite library", id.String())
	}
	return &library, nil
}

func (a *ase) GetAtlas(id uint64) (*Atlas, error) {
	a.mu.Lock()
	defer a.mu.Unlock()

	atlas, err := a.atlases.Get(slotmap.Unpack(id))
	if err != nil {
		return nil, errs.NotFound("aseprite atlas", fmt.Sprintf("%d", id)).Wrap(err)
	}

	return atlas, nil
}

func (a *ase) decodeAtlas(manifest resources.Resource) (*Atlas, error) {
	name := manifest.String()

	handle, err := a.res.GetHandle(manifest)
	if err != nil {
		return nil, fmt.Errorf("manifest %q: %w", name, err)
	}
	raw, err := a.dc.Get(handle)
	if err != nil {
		return nil, fmt.Errorf("manifest %q: %w", name, err)
	}

	var atlas Atlas
	if err := json.Unmarshal(raw, &atlas); err != nil {
		return nil, errs.Format(name, "invalid aseprite json").Wrap(err)
	}
	if dropped := atlas.DropMissingTags(); len(dropped) > 0 {
		a.logger.Warn("dropping missing tags", logger.Strings("tags", dropped))
	}
	if err := atlas.Validate(); err != nil {
		return nil, fmt.Errorf("manifest %q: %w", name, err)
	}
	return &atlas, nil
}

func (a *ase) isLibraryLoaded(bundle *resources.ResourceBundle) (bool, error) {
	for _, resource := range bundle.Resources {
		_, err := a.res.GetHandle(resource)
		if errs.IsNotFound(err) {
			return false, nil
		}
		if err != nil {
			return false, err
		}
	}
	return true, nil
}
