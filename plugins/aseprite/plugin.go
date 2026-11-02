package aseprite

import (
	"github.com/adm87/finch/pkg/errs"
	"github.com/adm87/finch/pkg/images"
	"github.com/adm87/finch/pkg/logger"
	"github.com/adm87/finch/pkg/resource"
	"github.com/adm87/finch/pkg/slotmap/store"
)

type Plugin struct {
	logger    logger.Logger
	resources *resource.Resources
	images    *images.Cache
	config    *Config
	libraries *store.Storage[Handle, LibraryInfo]
}

// NewPlugin checks the config. Libraries are read from res, from whichever
// source Load is given, and their images go in imageCache so the image
// renderer can draw them.
func NewPlugin(
	cfg *Config,
	res *resource.Resources,
	imageCache *images.Cache,
	log logger.Logger,
) (*Plugin, error) {
	if cfg == nil {
		return nil, errs.ArgumentNil("config")
	}
	if res == nil {
		return nil, errs.ArgumentNil("resources")
	}
	if err := cfg.Validate(); err != nil {
		return nil, err
	}
	return &Plugin{
		logger:    log.Named("aseprite"),
		resources: res,
		images:    imageCache,
		config:    cfg,
		libraries: store.New[Handle, LibraryInfo](store.WithKind(libraryKind)),
	}, nil
}
