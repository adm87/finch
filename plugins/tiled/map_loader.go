package tiled

import (
	"fmt"
	"image"
	"io/fs"
	"path"

	"github.com/adm87/finch/pkg/assets"
	"github.com/adm87/finch/pkg/errs"
	"github.com/adm87/finch/pkg/images"
	"github.com/adm87/finch/plugins/tiled/tmx"
)

// mapLoader builds one map's info. It follows the files the map refers to,
// loading each one just before it's parsed and recording the ones it loaded.
// Paths in the files are relative to the file they're in.
type mapLoader struct {
	plugin    *Plugin
	fsys      fs.FS
	tilesets  map[string]*tmx.Tileset // parsed once per map, by path
	templates map[string]struct{}
	owned     []assets.Asset
}

func (l *mapLoader) load(name string) (MapInfo, error) {
	tm, err := parse[tmx.Map](l, name)
	if err != nil {
		return MapInfo{}, err
	}
	if tm.Orientation != tmx.Orthogonal {
		return MapInfo{}, errs.NotSupported(fmt.Sprintf("%s maps", tm.Orientation))
	}
	dir := path.Dir(name)

	info := MapInfo{Map: tm, Tilesets: make([]TilesetInfo, 0, len(tm.Tilesets))}
	for _, mt := range tm.Tilesets {
		ts, err := l.mapTileset(mt, dir)
		if err != nil {
			return MapInfo{}, err
		}
		info.Tilesets = append(info.Tilesets, ts)
	}
	info.overdraw = overdraw(tm, info.Tilesets)

	for layer := range tm.Layers.All() {
		if layer.Image != nil {
			if err := l.require(imageRefs(dir, layer.Image.Image)...); err != nil {
				return MapInfo{}, err
			}
		}
		if layer.Objects == nil {
			continue
		}
		for _, obj := range layer.Objects.Objects {
			if obj.Template == "" {
				continue
			}
			if err := l.template(path.Join(dir, obj.Template)); err != nil {
				return MapInfo{}, err
			}
		}
	}

	layers, err := decodeLayers(tm)
	if err != nil {
		return MapInfo{}, err
	}
	info.Layers = layers
	return info, nil
}

// mapTileset builds a tileset the map uses, embedded or from a .tsx file.
func (l *mapLoader) mapTileset(mt tmx.MapTileset, dir string) (TilesetInfo, error) {
	ts := mt.Embedded
	if ts == nil {
		name := path.Join(dir, mt.Source)
		var err error
		if ts, err = l.tileset(name); err != nil {
			return TilesetInfo{}, err
		}
		dir = path.Dir(name)
	}

	info, err := l.tilesetImages(ts, dir)
	if err != nil {
		return TilesetInfo{}, fmt.Errorf("tileset %q: %w", ts.Name, err)
	}
	info.FirstGID = mt.FirstGID
	return info, nil
}

// tileset parses a .tsx file and loads its images.
func (l *mapLoader) tileset(name string) (*tmx.Tileset, error) {
	if ts, ok := l.tilesets[name]; ok {
		return ts, nil
	}
	ts, err := parse[tmx.Tileset](l, name)
	if err != nil {
		return nil, err
	}
	if err := l.require(tilesetImageRefs(ts, path.Dir(name))...); err != nil {
		return nil, err
	}
	l.tilesets[name] = ts
	return ts, nil
}

// template loads a template's tileset, so it's ready when objects are spawned.
func (l *mapLoader) template(name string) error {
	if _, ok := l.templates[name]; ok {
		return nil
	}
	tx, err := parse[tmx.Template](l, name)
	if err != nil {
		return err
	}
	l.templates[name] = struct{}{}
	if tx.Tileset == nil || tx.Tileset.Source == "" {
		return nil
	}
	_, err = l.tileset(path.Join(path.Dir(name), tx.Tileset.Source))
	return err
}

// tilesetImages loads the tileset's images and slices sheets into tiles.
func (l *mapLoader) tilesetImages(ts *tmx.Tileset, dir string) (TilesetInfo, error) {
	if err := l.require(tilesetImageRefs(ts, dir)...); err != nil {
		return TilesetInfo{}, err
	}

	info := TilesetInfo{Tileset: ts}
	if ts.Image != nil {
		handle, err := l.image(ts.Image, dir)
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

		handle, err := l.image(t.Image, dir)
		if err != nil {
			return TilesetInfo{}, fmt.Errorf("tile %d: %w", t.ID, err)
		}

		info.Tiles[t.ID] = handle
	}

	return info, nil
}

func (l *mapLoader) image(img *tmx.Image, dir string) (images.Handle, error) {
	if img.Source == "" {
		return 0, errs.NotSupported("images embedded as data")
	}
	name := path.Join(dir, img.Source)
	handle, exists := l.plugin.images.Lookup(name)
	if !exists {
		return 0, errs.NotFound("image", name)
	}
	return handle, nil
}

// require loads the assets that aren't loaded yet and owns them.
func (l *mapLoader) require(refs ...assets.Asset) error {
	var missing []assets.Asset
	for _, ref := range refs {
		if !l.plugin.assets.Loaded(ref) {
			missing = append(missing, ref)
		}
	}
	if len(missing) == 0 {
		return nil
	}

	err := l.plugin.assets.Load(l.fsys, missing...)

	// Load keeps going after a failure, so own whatever made it in.
	for _, ref := range missing {
		if l.plugin.assets.Loaded(ref) {
			l.owned = append(l.owned, ref)
		}
	}
	return err
}

// parse loads a Tiled file if needed and parses its bytes from the data cache.
func parse[T any, PT interface {
	*T
	tmx.Validator
}](l *mapLoader, name string) (*T, error) {
	if err := l.require(assets.Asset(name)); err != nil {
		return nil, err
	}
	handle, exists := l.plugin.data.Lookup(name)
	if !exists {
		return nil, errs.NotFound("data", name)
	}
	raw, err := l.plugin.data.Get(handle)
	if err != nil {
		return nil, fmt.Errorf("%q: %w", name, err)
	}
	value, err := tmx.Parse[T, PT](raw)
	if err != nil {
		return nil, errs.Format(name, "invalid tiled file").Wrap(err)
	}
	return value, nil
}

// tilesetImageRefs lists the image files a tileset draws from.
func tilesetImageRefs(ts *tmx.Tileset, dir string) []assets.Asset {
	refs := imageRefs(dir, ts.Image)
	for _, t := range ts.Tiles {
		refs = append(refs, imageRefs(dir, t.Image)...)
	}
	return refs
}

// imageRefs is the image's file, if it has one. Images embedded as data don't.
func imageRefs(dir string, img *tmx.Image) []assets.Asset {
	if img == nil || img.Source == "" {
		return nil
	}
	return []assets.Asset{assets.Asset(path.Join(dir, img.Source))}
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
