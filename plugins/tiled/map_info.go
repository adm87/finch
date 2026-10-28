package tiled

import (
	"image"
	"math"

	"github.com/hajimehoshi/ebiten/v2"

	"github.com/adm87/finch/pkg/assets"
	"github.com/adm87/finch/pkg/geom"
	"github.com/adm87/finch/pkg/images"
	"github.com/adm87/finch/plugins/tiled/tmx"
)

// MapInfo is a decoded map with everything needed to draw it.
type MapInfo struct {
	Map      *tmx.Map
	Tilesets []TilesetInfo // same order as Map.Tilesets
	Layers   []LayerInfo   // tile layers in draw order, groups flattened
	owned    []assets.Asset
	overdraw image.Point // how far tiles can reach past their cell, in pixels
}

// Bounds returns the rectangle that contains all the map's tile layers.
func (mi MapInfo) Bounds() geom.Rect {
	var bounds geom.Rect
	for _, layer := range mi.Layers {
		bounds = bounds.Union(layer.Bounds)
	}
	return bounds
}

// visibleTiles is the cells whose tiles can reach into view, which is in
// layer space.
func (mi MapInfo) visibleTiles(view geom.Rect) image.Rectangle {
	if view.Empty() {
		return image.Rectangle{}
	}
	tw, th := float64(mi.Map.TileWidth), float64(mi.Map.TileHeight)
	ox, oy := float64(mi.overdraw.X), float64(mi.overdraw.Y)
	minX, minY := view.Min()
	maxX, maxY := view.Max()
	return image.Rect(
		int(math.Floor((minX-ox)/tw)),
		int(math.Floor((minY-oy)/th)),
		int(math.Ceil((maxX+ox)/tw)),
		int(math.Ceil((maxY+oy)/th)),
	)
}

// LayerDepth finds the depth of any named layer, such as an empty "player"
// object layer to draw the player at.
func (mi MapInfo) LayerDepth(name string) (int32, bool) {
	var index int32
	for layer := range mi.Map.Layers.All() {
		if base := layer.Base(); base.Name == name {
			return layerDepth(base, index), true
		}
		index++
	}
	return 0, false
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
	Layer   *tmx.TileLayer
	GIDs    []tmx.GID
	Chunks  []ChunkInfo
	Color   ebiten.ColorScale // tint and opacity, with parent groups
	Bounds  geom.Rect         // map space, offsets included
	OffsetX float64           // with parent groups
	OffsetY float64
	Depth   int32 // added to the renderer's depth
	Visible bool  // with parent groups
}

type ChunkInfo struct {
	GIDs   []tmx.GID
	X      int
	Y      int
	Width  int
	Height int
}
