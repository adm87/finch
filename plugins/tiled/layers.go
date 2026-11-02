package tiled

import (
	"fmt"
	"strconv"

	"github.com/hajimehoshi/ebiten/v2"

	"github.com/adm87/finch/pkg/geom"
	"github.com/adm87/finch/plugins/tiled/tmx"
)

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
