package aseprite

import (
	"github.com/adm87/finch/pkg/assets"
	"github.com/adm87/finch/pkg/data"
	"github.com/adm87/finch/pkg/errs"
	"github.com/adm87/finch/pkg/images"
	"github.com/adm87/finch/pkg/logger"
	"github.com/adm87/finch/pkg/slotmap/store"
)

type Plugin struct {
	logger    logger.Logger
	assets    *assets.Assets
	data      *data.Cache
	images    *images.Cache
	config    *Config
	libraries *store.Storage[Handle, LibraryInfo]
}

// NewPlugin checks the config. Libraries are built from it on Load.
func NewPlugin(
	cfg *Config,
	assets *assets.Assets,
	dataCache *data.Cache,
	imageCache *images.Cache,
	log logger.Logger,
) (*Plugin, error) {
	if cfg == nil {
		return nil, errs.ArgumentNil("config")
	}
	if err := cfg.Validate(); err != nil {
		return nil, err
	}
	return &Plugin{
		logger:    log.Named("aseprite"),
		assets:    assets,
		data:      dataCache,
		images:    imageCache,
		config:    cfg,
		libraries: store.New[Handle, LibraryInfo](store.WithKind(libraryKind)),
	}, nil
}
