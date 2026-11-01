package aseprite

import (
	"io/fs"

	"github.com/adm87/finch/pkg/errs"
	"github.com/adm87/finch/pkg/images"
	"github.com/adm87/finch/pkg/logger"
	"github.com/adm87/finch/pkg/resource"
	"github.com/adm87/finch/pkg/slotmap/store"
)

// Sources resolves a source to the file system its files are read from.
type Sources interface {
	FS(src resource.Source) (fs.FS, error)
}

type Plugin struct {
	logger    logger.Logger
	sources   Sources
	images    *images.Cache
	config    *Config
	libraries *store.Storage[Handle, LibraryInfo]
}

// NewPlugin checks the config. Libraries are read from whichever source Load
// is given, and their images go in imageCache so the image renderer can draw them.
func NewPlugin(
	cfg *Config,
	sources Sources,
	imageCache *images.Cache,
	log logger.Logger,
) (*Plugin, error) {
	if cfg == nil {
		return nil, errs.ArgumentNil("config")
	}
	if sources == nil {
		return nil, errs.ArgumentNil("sources")
	}
	if err := cfg.Validate(); err != nil {
		return nil, err
	}
	return &Plugin{
		logger:    log.Named("aseprite"),
		sources:   sources,
		images:    imageCache,
		config:    cfg,
		libraries: store.New[Handle, LibraryInfo](store.WithKind(libraryKind)),
	}, nil
}
