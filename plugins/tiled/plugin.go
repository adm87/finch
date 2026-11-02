package tiled

import (
	"fmt"

	"github.com/adm87/finch/pkg/assets"
	"github.com/adm87/finch/pkg/data"
	"github.com/adm87/finch/pkg/errs"
	"github.com/adm87/finch/pkg/images"
	"github.com/adm87/finch/pkg/logger"
	"github.com/adm87/finch/pkg/rendering"
	"github.com/adm87/finch/pkg/slotmap/store"
)

// tiledTypes are stored as raw bytes in the data cache and parsed on Load.
var tiledTypes = []string{"tmx", "tsx", "tx"}

type Plugin struct {
	logger   logger.Logger
	assets   *assets.Assets
	data     *data.Cache
	images   *images.Cache
	config   *Config
	maps     *store.Storage[Handle, MapInfo]
	renderer rendering.AdapterID
}

// NewPlugin registers the Tiled importer and renderer.
func NewPlugin(
	cfg *Config,
	assets *assets.Assets,
	dataCache *data.Cache,
	imageCache *images.Cache,
	pipeline *rendering.Pipeline,
	log logger.Logger,
) (*Plugin, error) {
	if cfg == nil {
		return nil, errs.ArgumentNil("config")
	}
	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	p := &Plugin{
		logger: log.Named("tiled"),
		assets: assets,
		data:   dataCache,
		images: imageCache,
		config: cfg,
		maps:   store.New[Handle, MapInfo](store.WithKind(mapKind), store.WithCapacity(32)),
	}

	renderer, err := pipeline.RegisterRenderer(newTileMapRenderer(p))
	if err != nil {
		return nil, fmt.Errorf("register renderer: %w", err)
	}
	p.renderer = renderer

	if err := assets.RegisterImporter(data.NewImporter(dataCache), tiledTypes); err != nil {
		return nil, err
	}
	return p, nil
}
