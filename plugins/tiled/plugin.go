package tiled

import (
	"fmt"
	"io/fs"

	"github.com/adm87/finch/pkg/errs"
	"github.com/adm87/finch/pkg/images"
	"github.com/adm87/finch/pkg/logger"
	"github.com/adm87/finch/pkg/rendering"
	"github.com/adm87/finch/pkg/slotmap/store"
)

type Plugin struct {
	logger   logger.Logger
	fsys     fs.FS
	images   *images.Cache // tile images, only drawn by the plugin's renderer
	config   *Config
	maps     *store.Storage[Handle, MapInfo]
	renderer rendering.AdapterID
}

// NewPlugin registers the Tiled renderer. Maps and the files they refer to
// are read from fsys.
func NewPlugin(
	cfg *Config,
	fsys fs.FS,
	pipeline *rendering.Pipeline,
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

	p := &Plugin{
		logger: log.Named("tiled"),
		fsys:   fsys,
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
