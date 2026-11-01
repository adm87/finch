package aseprite

import (
	"io/fs"

	"github.com/adm87/finch/pkg/errs"
	"github.com/adm87/finch/pkg/images"
	"github.com/adm87/finch/pkg/logger"
	"github.com/adm87/finch/pkg/slotmap/store"
)

type Plugin struct {
	logger    logger.Logger
	fsys      fs.FS
	images    *images.Cache
	config    *Config
	libraries *store.Storage[Handle, LibraryInfo]
}

// NewPlugin checks the config. Libraries are read from fsys on Load, and
// their images go in imageCache so the image renderer can draw them.
func NewPlugin(
	cfg *Config,
	fsys fs.FS,
	imageCache *images.Cache,
	log logger.Logger,
) (*Plugin, error) {
	if cfg == nil {
		return nil, errs.ArgumentNil("config")
	}
	if fsys == nil {
		return nil, errs.ArgumentNil("fsys")
	}
	if err := cfg.Validate(); err != nil {
		return nil, err
	}
	return &Plugin{
		logger:    log.Named("aseprite"),
		fsys:      fsys,
		images:    imageCache,
		config:    cfg,
		libraries: store.New[Handle, LibraryInfo](store.WithKind(libraryKind)),
	}, nil
}
