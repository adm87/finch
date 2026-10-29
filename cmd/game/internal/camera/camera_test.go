package camera_test

import (
	"testing"

	"github.com/go-openapi/testify/require"

	"github.com/adm87/finch/cmd/game/internal/camera"
	"github.com/adm87/finch/pkg/geom"
	"github.com/adm87/finch/pkg/transform"
	"github.com/adm87/finch/pkg/world"
)

func TestMainFindsTheCamera(t *testing.T) {
	w := world.NewWorld()
	_, ok := camera.Main(w)
	require.False(t, ok)

	c := camera.NewMain(w)
	got, ok := camera.Main(w)
	require.True(t, ok)
	require.Equal(t, c.Entity(), got.Entity())
}

func TestZoomIsTheInverseOfScale(t *testing.T) {
	c := camera.NewMain(world.NewWorld())
	require.InDelta(t, 1.0, camera.Zoom(c), 1e-9)

	camera.SetZoom(c, 2)
	tr, _ := transform.GetTransform(c)
	sx, sy := tr.Scale()
	require.InDelta(t, 0.5, sx, 1e-9)
	require.InDelta(t, 0.5, sy, 1e-9)
	require.InDelta(t, 2.0, camera.Zoom(c), 1e-9)

	camera.ZoomBy(c, 1.5)
	require.InDelta(t, 3.0, camera.Zoom(c), 1e-9)
}

func TestSetZoomIgnoresNonPositive(t *testing.T) {
	c := camera.NewMain(world.NewWorld())
	camera.SetZoom(c, 2)

	camera.SetZoom(c, 0)
	camera.SetZoom(c, -1)
	require.InDelta(t, 2.0, camera.Zoom(c), 1e-9)
}

func TestFollowMovesTowardTheTarget(t *testing.T) {
	w := world.NewWorld()
	c := camera.NewMain(w)
	target := w.Create()
	tt, _ := transform.GetTransform(target)
	tt.SetPosition(100, 50)

	camera.Follow(0.5, c, target)

	tr, _ := transform.GetTransform(c)
	x, y := tr.Position()
	require.Greater(t, x, 0.0)
	require.LessOrEqual(t, x, 100.0)
	require.InDelta(t, x/2, y, 1e-9)
}

func TestClampKeepsTheViewInsideBounds(t *testing.T) {
	c := camera.NewMain(world.NewWorld())
	tr, _ := transform.GetTransform(c)
	tr.SetPosition(5, 5)

	view := geom.Rect{X: -15, Y: -10, Width: 40, Height: 30}
	camera.Clamp(c, view, geom.Rect{Width: 100, Height: 100})

	x, y := tr.Position()
	require.InDelta(t, 20.0, x, 1e-9)
	require.InDelta(t, 15.0, y, 1e-9)
}
