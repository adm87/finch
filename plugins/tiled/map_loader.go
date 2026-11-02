package tiled

import (
	"fmt"
	"image"
	"path"
	"strconv"

	"github.com/hajimehoshi/ebiten/v2"

	"github.com/adm87/finch/pkg/errs"
	"github.com/adm87/finch/pkg/geom"
	"github.com/adm87/finch/pkg/images"
	"github.com/adm87/finch/pkg/resource"
	"github.com/adm87/finch/plugins/tiled/tmx"
)

// mapLoader builds one map's info from one source. A map only names the files
// it needs once it's parsed, so it loads in stages: the map, its templates,
// its tilesets, then their images.
type mapLoader struct {
	plugin *Plugin
	src    resource.Source
}

func (l *mapLoader) load(name string) (MapInfo, error) {
	maps, err := loadAll(l, []string{name}, parse[tmx.Map])
	if err != nil {
		return MapInfo{}, err
	}
	tm := maps[name]
	if tm.Orientation != tmx.Orthogonal {
		return MapInfo{}, errs.NotSupported(fmt.Sprintf("%s maps", tm.Orientation))
	}

	refs := distinct(mapRefs(tm, path.Dir(name)))

	// Templates name tilesets of their own, so they're read first.
	templates, err := loadAll(l, withExt(refs, ".tx"), parse[tmx.Template])
	if err != nil {
		return MapInfo{}, err
	}
	tilesetNames := withExt(refs, ".tsx")
	for txName, tx := range templates {
		tilesetNames = append(tilesetNames, templateRefs(tx, path.Dir(txName))...)
	}

	tilesets, err := loadAll(l, distinct(tilesetNames), parse[tmx.Tileset])
	if err != nil {
		return MapInfo{}, err
	}
	for tsName, ts := range tilesets {
		tilesetRefs(ts, path.Dir(tsName))
	}

	// External tilesets cover the templates' too, so objects can be drawn later.
	var imageNames []string
	for _, ts := range tilesets {
		imageNames = append(imageNames, tilesetImageNames(ts)...)
	}
	for _, mt := range tm.Tilesets {
		if mt.Embedded != nil {
			imageNames = append(imageNames, tilesetImageNames(mt.Embedded)...)
		}
	}
	if err := images.Load(l.plugin.resources, l.plugin.images, l.src, distinct(imageNames)...); err != nil {
		return MapInfo{}, err
	}

	external := make(map[string]TilesetInfo, len(tilesets))
	for tsName, ts := range tilesets {
		info, err := l.tilesetImages(ts)
		if err != nil {
			return MapInfo{}, fmt.Errorf("tileset %q: %w", ts.Name, err)
		}
		external[tsName] = info
	}

	info := MapInfo{Map: tm, Tilesets: make([]TilesetInfo, 0, len(tm.Tilesets))}
	for _, mt := range tm.Tilesets {
		ts, ok := external[mt.Source]
		if mt.Embedded != nil {
			if ts, err = l.tilesetImages(mt.Embedded); err != nil {
				return MapInfo{}, fmt.Errorf("tileset %q: %w", mt.Embedded.Name, err)
			}
		} else if !ok {
			return MapInfo{}, errs.NotFound("tileset", mt.Source)
		}
		ts.FirstGID = mt.FirstGID
		info.Tilesets = append(info.Tilesets, ts)
	}
	info.overdraw = overdraw(tm, info.Tilesets)

	layers, err := decodeLayers(tm)
	if err != nil {
		return MapInfo{}, err
	}
	info.Layers = layers
	return info, nil
}

// tilesetImages looks up the tileset's images and slices sheets into tiles.
func (l *mapLoader) tilesetImages(ts *tmx.Tileset) (TilesetInfo, error) {
	info := TilesetInfo{Tileset: ts}
	if ts.Image != nil {
		handle, err := l.image(ts.Image)
		if err != nil {
			return TilesetInfo{}, err
		}

		img, err := l.plugin.images.Get(handle)
		if err != nil {
			return TilesetInfo{}, err
		}
		if err := l.plugin.images.SliceRects(handle, sheetRects(ts, img.Bounds())); err != nil {
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
	handle, ok := l.plugin.images.Lookup(resource.Key(l.src, img.Source))
	if !ok {
		return 0, errs.NotFound("image", img.Source)
	}
	return handle, nil
}

// loadAll reads and parses names from the loader's source with a throwaway
// resource.Loader, failing if any of them fails.
func loadAll[T any](l *mapLoader, names []string, decode func([]byte) (*T, error)) (map[string]*T, error) {
	ld := resource.NewLoader(l.plugin.resources, decode)
	ld.Queue(l.src, names...)
	results := ld.Load()
	if err := resource.Err(results); err != nil {
		return nil, err
	}
	out := make(map[string]*T, len(results))
	for _, r := range results {
		out[r.Name] = r.Value
	}
	return out, nil
}

func parse[T any, PT interface {
	*T
	tmx.Validator
}](raw []byte) (*T, error) {
	value, err := tmx.Parse[T, PT](raw)
	if err != nil {
		return nil, errs.Format("tiled file", "invalid").Wrap(err)
	}
	return value, nil
}

// tilesetImageNames lists the image files a tileset draws from. Images
// embedded as data have no file and are reported when the tileset is built.
func tilesetImageNames(ts *tmx.Tileset) []string {
	var names []string
	if ts.Image != nil && ts.Image.Source != "" {
		names = append(names, ts.Image.Source)
	}
	for _, t := range ts.Tiles {
		if t.Image != nil && t.Image.Source != "" {
			names = append(names, t.Image.Source)
		}
	}
	return names
}

func withExt(names []string, ext string) []string {
	var out []string
	for _, name := range names {
		if path.Ext(name) == ext {
			out = append(out, name)
		}
	}
	return out
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
