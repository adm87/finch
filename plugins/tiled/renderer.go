package tiled

import (
	"image"
	"math"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/yohamta/donburi"

	"github.com/adm87/finch/pkg/geom"
	"github.com/adm87/finch/pkg/images"
	"github.com/adm87/finch/pkg/rendering"
	"github.com/adm87/finch/plugins/tiled/tmx"
)

// TileMapRenderer draws each tile layer of a map as its own job, so other
// entities can draw between layers.
type TileMapRenderer struct {
	maps   *Maps
	images *images.Cache
	opts   ebiten.DrawImageOptions
}

func NewTileMapRenderer(maps *Maps, images *images.Cache) *TileMapRenderer {
	return &TileMapRenderer{
		maps:   maps,
		images: images,
	}
}

func (tm *TileMapRenderer) Jobs(candidate *rendering.Candidate, jobs []rendering.Job) []rendering.Job {
	info, ok := tm.mapInfo(candidate.Entry)
	if !ok {
		return jobs
	}

	geoM := candidate.Transform.GeoM()
	for i, layer := range info.Layers {
		if !layer.Visible {
			continue
		}
		job := candidate.Job()
		job.Pass = i
		job.Depth += layer.Depth
		job.Bounds = geom.TransformRect(geoM, layer.Bounds)
		jobs = append(jobs, job)
	}
	return jobs
}

func (tm *TileMapRenderer) Render(target *ebiten.Image, job *rendering.Job, viewRect geom.Rect, viewMatrix geom.Matrix) {
	info, ok := tm.mapInfo(job.Entry)
	if !ok || job.Pass >= len(info.Layers) {
		return
	}
	layer := &info.Layers[job.Pass]

	// Tiles are placed in layer space, then moved like the entity.
	var toWorld geom.Matrix
	toWorld.Translate(layer.OffsetX, layer.OffsetY)
	toWorld.Concat(job.Transform.GeoM())
	toTarget := toWorld
	toTarget.Concat(viewMatrix)

	visible := info.visibleTiles(layerView(toWorld, viewRect))
	tm.opts.ColorScale = layer.Color

	if layer.Chunks == nil {
		width := layer.Layer.Width
		visible = visible.Intersect(image.Rect(0, 0, width, layer.Layer.Height))
		for y := visible.Min.Y; y < visible.Max.Y; y++ {
			for x := visible.Min.X; x < visible.Max.X; x++ {
				tm.drawTile(target, &info, layer.GIDs[y*width+x], x, y, toTarget)
			}
		}
		return
	}

	for _, c := range layer.Chunks {
		area := visible.Intersect(image.Rect(c.X, c.Y, c.X+c.Width, c.Y+c.Height))
		for y := area.Min.Y; y < area.Max.Y; y++ {
			for x := area.Min.X; x < area.Max.X; x++ {
				tm.drawTile(target, &info, c.GIDs[(y-c.Y)*c.Width+(x-c.X)], x, y, toTarget)
			}
		}
	}
}

func (tm *TileMapRenderer) mapInfo(entry *donburi.Entry) (MapInfo, bool) {
	m, ok := GetTileMap(entry)
	if !ok {
		return MapInfo{}, false
	}
	info, err := tm.maps.Map(m.Handle())
	return info, err == nil
}

// drawTile draws a tile in cell x, y. Tiles bigger than a cell keep to the
// cell's bottom left, like Tiled draws them.
func (tm *TileMapRenderer) drawTile(target *ebiten.Image, info *MapInfo, gid tmx.GID, x, y int, toTarget geom.Matrix) {
	ts, id, ok := info.Tileset(gid)
	if !ok {
		return
	}
	handle, frameIndex, ok := ts.TileImage(id)
	if !ok {
		return
	}
	frame, err := tm.images.Frame(handle, frameIndex)
	if err != nil {
		return
	}

	w, h := float64(frame.Size.X), float64(frame.Size.Y)
	tm.opts.GeoM.Reset()
	flipTile(&tm.opts.GeoM, gid, w, h)
	if gid.FlippedDiagonally() {
		h = w // it's drawn on its side
	}

	var offsetX, offsetY float64
	if to := ts.Tileset.TileOffset; to != nil {
		offsetX, offsetY = float64(to.X), float64(to.Y)
	}
	tm.opts.GeoM.Translate(
		float64(x*info.Map.TileWidth+frame.Offset.X)+offsetX,
		float64((y+1)*info.Map.TileHeight+frame.Offset.Y)-h+offsetY,
	)
	tm.opts.GeoM.Concat(toTarget)

	target.DrawImage(frame.Image, &tm.opts)
}

// flipTile applies the gid's flips to a w×h tile, keeping it at the origin.
// Tiled flips diagonally first, then horizontally, then vertically.
func flipTile(m *geom.Matrix, gid tmx.GID, w, h float64) {
	if !gid.FlippedDiagonally() && !gid.FlippedHorizontally() && !gid.FlippedVertically() {
		return
	}

	m.Translate(-w/2, -h/2)
	if gid.FlippedDiagonally() {
		m.Rotate(math.Pi / 2) // with the x flip below, this swaps x and y
		m.Scale(-1, 1)
		w, h = h, w
	}
	if gid.FlippedHorizontally() {
		m.Scale(-1, 1)
	}
	if gid.FlippedVertically() {
		m.Scale(1, -1)
	}
	m.Translate(w/2, h/2)
}

// layerView is the world-space view in layer space.
func layerView(toWorld geom.Matrix, viewRect geom.Rect) geom.Rect {
	if !toWorld.IsInvertible() {
		return geom.Rect{}
	}
	toWorld.Invert()
	return geom.TransformRect(toWorld, viewRect)
}
