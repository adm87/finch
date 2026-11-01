package tiled

import (
	"fmt"
	"io/fs"

	"github.com/adm87/finch/pkg/errs"
	"github.com/adm87/finch/pkg/images"
	"github.com/adm87/finch/pkg/logger"
	"github.com/adm87/finch/pkg/rendering"
	"github.com/adm87/finch/pkg/resource"
	"github.com/adm87/finch/pkg/slotmap/store"
)

// Sources resolves a source to the file system its files are read from.
type Sources interface {
	FS(src resource.Source) (fs.FS, error)
}

type Plugin struct {
	logger   logger.Logger
	sources  Sources
	images   *images.Cache // tile images, only drawn by the plugin's renderer
	config   *Config
	maps     *store.Storage[Handle, MapInfo]
	renderer rendering.AdapterID
}

// NewPlugin registers the Tiled renderer. Maps and the files they refer to
// are read from whichever source Load is given.
func NewPlugin(
	cfg *Config,
	sources Sources,
	pipeline *rendering.Pipeline,
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

	p := &Plugin{
		logger: log.Named("tiled"),
		sources: sources,
		images: images.NewCache(),
		config: cfg,
		maps:   store.New[Handle, MapInfo](store.WithKind(mapKind), store.WithCapacity(32)),
	}

	renderer, err := pipeline.RegisterRenderer(newTileMapRenderer(p))
	if err != nil {
		return nil, fmt.Errorf("register renderer: %w", err)
	}
	p.renderer = renderer
	return p, nil
}
