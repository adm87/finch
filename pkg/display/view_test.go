package display_test

import (
	"testing"

	"github.com/go-openapi/testify/require"
	"github.com/yohamta/donburi"

	"github.com/adm87/finch/pkg/display"
	"github.com/adm87/finch/pkg/transform"
)

func TestViewConversionsAreInverses(t *testing.T) {
	screen := display.NewScreen(320, 180)
	screen.Fit(1280, 800)
	require.NotZero(t, screen.SafeArea().X)
	require.NotEqual(t, 1.0, screen.Scale())

	w := donburi.NewWorld()
	camera := w.Entry(w.Create(transform.TransformComponent, transform.BoundsComponent))
	tr, _ := transform.GetTransform(camera)
	tr.SetPosition(50, -20)
	tr.SetScale(2, 2)

	view := display.NewViewport(screen)
	view.SetCamera(camera)

	for _, p := range [][2]float64{{0, 0}, {123, 45}, {-30, 200}, {1280, 800}} {
		wx, wy := view.ScreenToWorld(p[0], p[1])
		sx, sy := view.WorldToScreen(wx, wy)
		require.InDelta(t, p[0], sx, 1e-9)
		require.InDelta(t, p[1], sy, 1e-9)
	}
}
