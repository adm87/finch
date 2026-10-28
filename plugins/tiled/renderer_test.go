package tiled_test

import (
	"testing"

	"github.com/go-openapi/testify/require"
	"github.com/yohamta/donburi"

	"github.com/adm87/finch/pkg/geom"
	"github.com/adm87/finch/pkg/images"
	"github.com/adm87/finch/pkg/rendering"
	"github.com/adm87/finch/pkg/transform"
	"github.com/adm87/finch/plugins/tiled"
)

func TestLayersFoldInTheirGroups(t *testing.T) {
	maps, _, _ := newMaps(t)
	handle, err := maps.CreateMap("layered", tiled.LoadModeDynamic)
	require.NoError(t, err)
	info, err := maps.Map(handle)
	require.NoError(t, err)

	require.Len(t, info.Layers, 3)
	ground, roofs, hidden := info.Layers[0], info.Layers[1], info.Layers[2]

	require.Equal(t, int32(0), ground.Depth)
	require.True(t, ground.Visible)
	require.Equal(t, geom.Rect{Width: 32, Height: 32}, ground.Bounds)

	// The group's offset and opacity carry into its layers.
	require.Equal(t, int32(50), roofs.Depth)
	require.Equal(t, 10.0, roofs.OffsetX)
	require.Equal(t, 4.0, roofs.OffsetY)
	require.Equal(t, geom.Rect{X: 10, Y: 4, Width: 32, Height: 32}, roofs.Bounds)
	require.InDelta(t, 0.5, roofs.Color.A(), 1e-6)

	require.Equal(t, int32(4), hidden.Depth) // ground, player, above, roofs, hidden
	require.False(t, hidden.Visible)

	require.Equal(t, geom.Rect{Width: 42, Height: 36}, info.Bounds())

	depth, ok := info.LayerDepth("player")
	require.True(t, ok)
	require.Equal(t, int32(1), depth)
	depth, ok = info.LayerDepth("shadow")
	require.True(t, ok)
	require.Equal(t, int32(-3), depth)
	_, ok = info.LayerDepth("nope")
	require.False(t, ok)
}

func TestRendererMakesAJobPerVisibleLayer(t *testing.T) {
	maps, _, _ := newMaps(t)
	handle, err := maps.CreateMap("layered", tiled.LoadModeDynamic)
	require.NoError(t, err)

	w := donburi.NewWorld()
	e := w.Entry(w.Create(tiled.TileMapComponent, transform.TransformComponent))
	tm, _ := tiled.GetTileMap(e)
	tm.SetHandle(handle)
	tf, _ := transform.GetTransform(e)
	tf.SetPosition(100, 0)

	r := &rendering.Renderer{}
	r.SetLayer(2)
	r.SetDepth(10)

	renderer := tiled.NewTileMapRenderer(maps, images.NewCache())
	jobs := renderer.Jobs(&rendering.Candidate{Entry: e, Renderer: r, Transform: tf}, nil)

	require.Len(t, jobs, 2) // the hidden layer is skipped
	require.Equal(t, 0, jobs[0].Pass)
	require.Equal(t, int32(10), jobs[0].Depth)
	require.Equal(t, int16(2), jobs[0].Layer)
	require.Equal(t, geom.Rect{X: 100, Width: 32, Height: 32}, jobs[0].Bounds)

	require.Equal(t, 1, jobs[1].Pass)
	require.Equal(t, int32(60), jobs[1].Depth)
	require.Equal(t, geom.Rect{X: 110, Y: 4, Width: 32, Height: 32}, jobs[1].Bounds)
}

func TestRendererSkipsEntitiesWithoutAMap(t *testing.T) {
	maps, _, _ := newMaps(t)
	w := donburi.NewWorld()
	e := w.Entry(w.Create(transform.TransformComponent))
	tf, _ := transform.GetTransform(e)

	renderer := tiled.NewTileMapRenderer(maps, images.NewCache())
	jobs := renderer.Jobs(&rendering.Candidate{Entry: e, Renderer: &rendering.Renderer{}, Transform: tf}, nil)
	require.Empty(t, jobs)
}
