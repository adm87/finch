package tiled

import (
	"errors"
	"fmt"
	"image"
	"io/fs"
	"path"
	"strconv"

	"github.com/adm87/finch/pkg/assets"
	"github.com/adm87/finch/pkg/data"
	"github.com/adm87/finch/pkg/errs"
	"github.com/adm87/finch/pkg/images"
	"github.com/adm87/finch/pkg/logger"
	"github.com/adm87/finch/pkg/slotmap"
	"github.com/adm87/finch/plugins/tiled/tmx"
)

const mapKind = "tiled map"

// Handle uniquely identifies a Tiled map.
type Handle uint64

// LoadMode controls whether CreateMap may load missing files.
type LoadMode uint8

const (
	// LoadModeRequired fails if any of the map's files isn't already loaded.
	LoadModeRequired LoadMode = iota
	// LoadModeDynamic loads the map's files that aren't loaded.
	LoadModeDynamic
)

// UnloadMode controls whether ClearMap unloads the map's files.
type UnloadMode int

const (
	// UnloadModeKeep leaves the files loaded.
	UnloadModeKeep UnloadMode = iota
	// UnloadModeIfOwned unloads the files CreateMap loaded.
	UnloadModeIfOwned
)

// MapInfo is a decoded map with everything needed to draw it.
type MapInfo struct {
	Map      *tmx.Map
	Tilesets []TilesetInfo // same order as Map.Tilesets
	Layers   []LayerInfo   // tile layers in draw order, groups flattened
	owned    []assets.Asset
}

// Tileset finds the tileset a tile belongs to and the tile's local ID.
func (mi MapInfo) Tileset(gid tmx.GID) (*TilesetInfo, int, bool) {
	id := int(gid.ID())
	if id == 0 {
		return nil, 0, false
	}
	for i := len(mi.Tilesets) - 1; i >= 0; i-- {
		if ts := &mi.Tilesets[i]; ts.FirstGID <= id {
			return ts, id - ts.FirstGID, true
		}
	}
	return nil, 0, false
}

// TilesetInfo is a tileset and its images. A sheet is sliced so a tile's local
// ID is its frame index. A collection has one image per tile instead.
type TilesetInfo struct {
	Tileset  *tmx.Tileset
	Tiles    []images.Handle // collections only, indexed by local ID, 0 for none
	Image    images.Handle   // sheets only
	FirstGID int
}

// TileImage returns the image and frame to draw a tile with.
func (ti *TilesetInfo) TileImage(id int) (images.Handle, int, bool) {
	if ti.Tiles == nil {
		return ti.Image, id, ti.Image != 0
	}
	if id < 0 || id >= len(ti.Tiles) || ti.Tiles[id] == 0 {
		return 0, 0, false
	}
	return ti.Tiles[id], 0, true
}

// LayerInfo is a tile layer with its tiles decoded. Finite maps fill GIDs and
// infinite maps fill Chunks.
type LayerInfo struct {
	Layer  *tmx.TileLayer
	GIDs   []tmx.GID
	Chunks []ChunkInfo
}

type ChunkInfo struct {
	GIDs   []tmx.GID
	X      int
	Y      int
	Width  int
	Height int
}

type Maps struct {
	logger     logger.Logger
	assets     *assets.Assets
	data       *data.Cache
	images     *images.Cache
	config     *Config
	maps       *slotmap.SlotMap[MapInfo]
	mapsByName map[string]Handle
}

func NewMaps(as *assets.Assets, data *data.Cache, images *images.Cache, cfg *Config, log logger.Logger) *Maps {
	return &Maps{
		logger:     log.Named("tiled"),
		assets:     as,
		data:       data,
		images:     images,
		config:     cfg,
		maps:       slotmap.New[MapInfo](32),
		mapsByName: make(map[string]Handle),
	}
}

func (m *Maps) CreateMap(name string, mode LoadMode) (Handle, error) {
	if handle, exists := m.mapsByName[name]; exists {
		return handle, nil
	}

	cfg, err := m.getMapConfig(name)
	if err != nil {
		return 0, err
	}
	fsys, err := m.assets.FileSystem(cfg.FileSystem)
	if err != nil {
		return 0, fmt.Errorf("map %q: %w", name, err)
	}

	l := &mapLoader{maps: m, fsys: fsys, mode: mode, tilesets: make(map[assets.Asset]*tmx.Tileset)}
	info, err := l.load(cfg.TMX)
	if err != nil {
		return 0, fmt.Errorf("map %q: %w", name, errors.Join(err, m.unload(l.owned)))
	}
	info.owned = l.owned

	handle := Handle(m.maps.Insert(info).Pack())
	m.mapsByName[name] = handle

	return handle, nil
}

func (m *Maps) ClearMap(name string, mode UnloadMode) error {
	handle, exists := m.mapsByName[name]
	if !exists {
		return errs.NotFound(mapKind, name)
	}

	info, err := m.maps.Delete(slotmap.Unpack(uint64(handle)))
	delete(m.mapsByName, name)
	if err != nil {
		return fmt.Errorf("clear map %q: %w", name, err)
	}

	if mode != UnloadModeIfOwned {
		return nil
	}
	if err := m.unload(info.owned); err != nil {
		return fmt.Errorf("unload map %q: %w", name, err)
	}
	return nil
}

func (m *Maps) Map(handle Handle) (MapInfo, error) {
	info, err := m.maps.Get(slotmap.Unpack(uint64(handle)))
	if err != nil {
		return MapInfo{}, errs.NotFound(mapKind, strconv.FormatUint(uint64(handle), 10)).Wrap(err)
	}
	return info, nil
}

func (m *Maps) getMapConfig(name string) (*MapConfig, error) {
	if name == "" {
		return nil, errs.Argument("name", "cannot be empty")
	}
	cfg, exists := m.config.Maps[name]
	if !exists {
		return nil, errs.NotFound(mapKind, name)
	}
	return &cfg, nil
}

func (m *Maps) unload(owned []assets.Asset) error {
	if len(owned) == 0 {
		return nil
	}
	return m.assets.Unload(owned...)
}

// mapLoader builds one map's info, recording the files it had to load.
type mapLoader struct {
	maps     *Maps
	fsys     fs.FS
	tilesets map[assets.Asset]*tmx.Tileset // parsed once per map
	owned    []assets.Asset
	mode     LoadMode
}

func (l *mapLoader) load(asset assets.Asset) (MapInfo, error) {
	tm, err := parseAsset[tmx.Map](l, asset)
	if err != nil {
		return MapInfo{}, err
	}

	refs := distinct(mapRefs(tm, path.Dir(asset.String())))
	if err := l.require(refs...); err != nil {
		return MapInfo{}, err
	}

	info := MapInfo{Map: tm, Tilesets: make([]TilesetInfo, 0, len(tm.Tilesets))}
	for _, mt := range tm.Tilesets {
		ts, err := l.mapTileset(mt)
		if err != nil {
			return MapInfo{}, err
		}
		info.Tilesets = append(info.Tilesets, ts)
	}

	for _, ref := range refs {
		if ref.Type() != "tx" {
			continue
		}
		if err := l.template(ref); err != nil {
			return MapInfo{}, err
		}
	}

	layers, err := decodeLayers(tm)
	if err != nil {
		return MapInfo{}, err
	}
	info.Layers = layers
	return info, nil
}

func (l *mapLoader) mapTileset(mt tmx.MapTileset) (TilesetInfo, error) {
	ts := mt.Embedded
	if ts == nil {
		var err error
		if ts, err = l.tileset(assets.Asset(mt.Source)); err != nil {
			return TilesetInfo{}, err
		}
	}

	info, err := l.tilesetImages(ts)
	if err != nil {
		return TilesetInfo{}, fmt.Errorf("tileset %q: %w", ts.Name, err)
	}
	info.FirstGID = mt.FirstGID
	return info, nil
}

// tileset parses a .tsx file and loads its images.
func (l *mapLoader) tileset(asset assets.Asset) (*tmx.Tileset, error) {
	if ts, ok := l.tilesets[asset]; ok {
		return ts, nil
	}
	ts, err := parseAsset[tmx.Tileset](l, asset)
	if err != nil {
		return nil, err
	}
	if err := l.require(tilesetRefs(ts, path.Dir(asset.String()))...); err != nil {
		return nil, err
	}
	l.tilesets[asset] = ts
	return ts, nil
}

// template loads a template's tileset, so it's ready when objects are spawned.
func (l *mapLoader) template(asset assets.Asset) error {
	tx, err := parseAsset[tmx.Template](l, asset)
	if err != nil {
		return err
	}
	for _, ref := range templateRefs(tx, path.Dir(asset.String())) {
		if _, err := l.tileset(ref); err != nil {
			return err
		}
	}
	return nil
}

func (l *mapLoader) tilesetImages(ts *tmx.Tileset) (TilesetInfo, error) {
	info := TilesetInfo{Tileset: ts}
	if ts.Image != nil {
		handle, err := l.image(ts.Image)
		if err != nil {
			return TilesetInfo{}, err
		}
		img, err := l.maps.images.Get(handle)
		if err != nil {
			return TilesetInfo{}, err
		}
		if err := l.maps.images.SliceRects(handle, sheetRects(ts, img.Bounds())); err != nil {
			return TilesetInfo{}, fmt.Errorf("slice image %q: %w", ts.Image.Source, err)
		}
		info.Image = handle
		return info, nil
	}

	size := 0
	for _, t := range ts.Tiles {
		size = max(size, t.ID+1)
	}
	info.Tiles = make([]images.Handle, size)
	for _, t := range ts.Tiles {
		if t.Image == nil {
			continue
		}
		handle, err := l.image(t.Image)
		if err != nil {
			return TilesetInfo{}, fmt.Errorf("tile %d: %w", t.ID, err)
		}
		info.Tiles[t.ID] = handle
	}
	return info, nil
}

func (l *mapLoader) image(img *tmx.Image) (images.Handle, error) {
	if img.Source == "" {
		return 0, errs.NotSupported("images embedded as data")
	}
	handle, err := l.maps.assets.Handle(assets.Asset(img.Source))
	if err != nil {
		return 0, fmt.Errorf("image %q: %w", img.Source, err)
	}
	return images.Handle(handle), nil
}

// require loads the assets that aren't loaded, or fails in LoadModeRequired.
func (l *mapLoader) require(refs ...assets.Asset) error {
	var missing []assets.Asset
	for _, ref := range refs {
		if _, err := l.maps.assets.Handle(ref); err != nil {
			missing = append(missing, ref)
		}
	}
	if len(missing) == 0 {
		return nil
	}
	if l.mode == LoadModeRequired {
		return errs.InvalidOperation(fmt.Sprintf("%v not loaded and mode is LoadModeRequired", missing))
	}

	err := l.maps.assets.Load(l.fsys, missing...)

	// Load keeps going after a failure, so own whatever made it in.
	for _, ref := range missing {
		if _, herr := l.maps.assets.Handle(ref); herr == nil {
			l.owned = append(l.owned, ref)
		}
	}
	return err
}

// parseAsset loads asset if needed and parses its bytes from the data cache.
func parseAsset[T tmx.Validator[T]](l *mapLoader, asset assets.Asset) (*T, error) {
	if err := l.require(asset); err != nil {
		return nil, err
	}
	handle, err := l.maps.assets.Handle(asset)
	if err != nil {
		return nil, err
	}
	raw, err := l.maps.data.Get(data.Handle(handle))
	if err != nil {
		return nil, fmt.Errorf("%q: %w", asset, err)
	}
	value, err := tmx.Parse[T](raw)
	if err != nil {
		return nil, errs.Format(asset.String(), "invalid tiled file").Wrap(err)
	}
	return value, nil
}

// sheetRects cuts a tileset image so frame i is local tile i. Older files
// leave out columns and tilecount, so they fall back to what fits.
func sheetRects(ts *tmx.Tileset, bounds image.Rectangle) []image.Rectangle {
	stepX, stepY := ts.TileWidth+ts.Spacing, ts.TileHeight+ts.Spacing

	columns := ts.Columns
	if columns <= 0 {
		columns = (bounds.Dx() - 2*ts.Margin + ts.Spacing) / stepX
	}
	count := ts.TileCount
	if count <= 0 {
		count = columns * ((bounds.Dy() - 2*ts.Margin + ts.Spacing) / stepY)
	}

	rects := make([]image.Rectangle, count)
	for id := range rects {
		x := bounds.Min.X + ts.Margin + (id%columns)*stepX
		y := bounds.Min.Y + ts.Margin + (id/columns)*stepY
		rects[id] = image.Rect(x, y, x+ts.TileWidth, y+ts.TileHeight)
	}
	return rects
}

func decodeLayers(tm *tmx.Map) ([]LayerInfo, error) {
	var layers []LayerInfo
	for layer := range tm.Layers.All() {
		if layer.Tiles == nil {
			continue
		}
		info, err := decodeLayer(layer.Tiles)
		if err != nil {
			return nil, fmt.Errorf("layer %q: %w", layer.Tiles.Name, err)
		}
		layers = append(layers, info)
	}
	return layers, nil
}

func decodeLayer(tl *tmx.TileLayer) (LayerInfo, error) {
	info := LayerInfo{Layer: tl}
	if len(tl.Data.Chunks) == 0 {
		gids, err := tl.Data.GIDs()
		info.GIDs = gids
		return info, err
	}

	info.Chunks = make([]ChunkInfo, len(tl.Data.Chunks))
	for i, c := range tl.Data.Chunks {
		gids, err := tl.Data.ChunkGIDs(c)
		if err != nil {
			return LayerInfo{}, fmt.Errorf("chunk %d,%d: %w", c.X, c.Y, err)
		}
		info.Chunks[i] = ChunkInfo{GIDs: gids, X: c.X, Y: c.Y, Width: c.Width, Height: c.Height}
	}
	return info, nil
}
