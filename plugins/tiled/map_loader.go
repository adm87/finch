package tiled

import (
	"fmt"
	"image"
	"io/fs"
	"path"
	"strconv"

	"github.com/hajimehoshi/ebiten/v2"

	"github.com/adm87/finch/pkg/assets"
	"github.com/adm87/finch/pkg/data"
	"github.com/adm87/finch/pkg/errs"
	"github.com/adm87/finch/pkg/geom"
	"github.com/adm87/finch/pkg/images"
	"github.com/adm87/finch/plugins/tiled/tmx"
)

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
	if tm.Orientation != tmx.Orthogonal {
		return MapInfo{}, errs.NotSupported(fmt.Sprintf("%s maps", tm.Orientation))
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
	info.overdraw = overdraw(tm, info.Tilesets)

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
func parseAsset[T any, PT interface {
	*T
	tmx.Validator
}](l *mapLoader, asset assets.Asset) (*T, error) {
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
	value, err := tmx.Parse[T, PT](raw)
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

// overdraw is how far any tile can reach past its cell. Tiles bigger than
// the map's cells stick out, flips can turn them, and offsets move them.
func overdraw(tm *tmx.Map, tilesets []TilesetInfo) image.Point {
	var reach image.Point
	for _, info := range tilesets {
		ts := info.Tileset
		size := max(ts.TileWidth, ts.TileHeight)
		x, y := max(0, size-tm.TileWidth), max(0, size-tm.TileHeight)
		if o := ts.TileOffset; o != nil {
			x += abs(o.X)
			y += abs(o.Y)
		}
		reach.X, reach.Y = max(reach.X, x), max(reach.Y, y)
	}
	return reach
}

func abs(v int) int {
	if v < 0 {
		return -v
	}
	return v
}

// layerState is what a group passes down to the layers inside it.
type layerState struct {
	color   ebiten.ColorScale
	offsetX float64
	offsetY float64
	visible bool
}

func (s layerState) with(b *tmx.LayerBase) layerState {
	s.offsetX += b.OffsetX
	s.offsetY += b.OffsetY
	s.visible = s.visible && b.Visible
	s.color.ScaleAlpha(float32(b.Opacity))
	if b.TintColor != nil {
		s.color.ScaleWithColor(b.TintColor)
	}
	return s
}

// layerDecoder flattens a map's tile layers, folding group offsets, tint,
// opacity and visibility into each one.
type layerDecoder struct {
	tm     *tmx.Map
	layers []LayerInfo
	index  int32 // position in Layers.All, for depth
}

func decodeLayers(tm *tmx.Map) ([]LayerInfo, error) {
	d := &layerDecoder{tm: tm}
	if err := d.walk(tm.Layers, layerState{visible: true}); err != nil {
		return nil, err
	}
	return d.layers, nil
}

func (d *layerDecoder) walk(layers tmx.Layers, parent layerState) error {
	for _, layer := range layers {
		base := layer.Base()
		depth := layerDepth(base, d.index)
		d.index++
		state := parent.with(base)

		if layer.Group != nil {
			if err := d.walk(layer.Group.Layers, state); err != nil {
				return err
			}
			continue
		}
		if layer.Tiles == nil {
			continue
		}

		info, err := decodeLayer(layer.Tiles)
		if err != nil {
			return fmt.Errorf("layer %q: %w", base.Name, err)
		}
		info.Color = state.color
		info.OffsetX, info.OffsetY = state.offsetX, state.offsetY
		info.Visible = state.visible
		info.Depth = depth
		info.Bounds = d.bounds(info)
		d.layers = append(d.layers, info)
	}
	return nil
}

// bounds is the layer's area in pixels. Infinite layers cover their chunks.
func (d *layerDecoder) bounds(info LayerInfo) geom.Rect {
	tw, th := float64(d.tm.TileWidth), float64(d.tm.TileHeight)
	tiles := func(x, y, w, h int) geom.Rect {
		return geom.Rect{
			X:      float64(x)*tw + info.OffsetX,
			Y:      float64(y)*th + info.OffsetY,
			Width:  float64(w) * tw,
			Height: float64(h) * th,
		}
	}

	if info.Chunks == nil {
		return tiles(0, 0, info.Layer.Width, info.Layer.Height)
	}
	var bounds geom.Rect
	for _, c := range info.Chunks {
		bounds = bounds.Union(tiles(c.X, c.Y, c.Width, c.Height))
	}
	return bounds
}

// layerDepth is a "depth" int property if the layer has one, or its position
// among all the map's layers, object layers included.
func layerDepth(b *tmx.LayerBase, index int32) int32 {
	if p, ok := b.Properties.Get("depth"); ok {
		if depth, err := strconv.ParseInt(p.Value, 10, 32); err == nil {
			return int32(depth)
		}
	}
	return index
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
