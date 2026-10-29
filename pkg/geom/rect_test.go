package geom_test

import (
	"testing"

	"github.com/go-openapi/testify/require"

	"github.com/adm87/finch/pkg/geom"
)

func TestRectUnion(t *testing.T) {
	a := geom.Rect{X: 0, Y: 0, Width: 10, Height: 10}
	b := geom.Rect{X: 5, Y: -5, Width: 10, Height: 10}

	require.Equal(t, geom.Rect{X: 0, Y: -5, Width: 15, Height: 15}, a.Union(b))
	require.Equal(t, a.Union(b), b.Union(a))
	require.Equal(t, a, a.Union(geom.Rect{X: 2, Y: 2, Width: 1, Height: 1}))

	// Empty rects add nothing, wherever they are.
	far := geom.Rect{X: 32, Y: 32, Width: 8, Height: 8}
	require.Equal(t, far, geom.Rect{}.Union(far))
	require.Equal(t, far, far.Union(geom.Rect{}))
	require.Equal(t, far, far.Union(geom.Rect{X: -100, Y: -100}))
}

func TestRectEmpty(t *testing.T) {
	require.True(t, geom.Rect{}.Empty())
	require.True(t, geom.Rect{Width: 5}.Empty())
	require.True(t, geom.Rect{Width: 5, Height: -1}.Empty())
	require.False(t, geom.Rect{Width: 1, Height: 1}.Empty())
}

func TestRectClampPoint(t *testing.T) {
	r := geom.Rect{X: 10, Y: 20, Width: 100, Height: 50}

	x, y := r.ClampPoint(50, 40)
	require.Equal(t, [2]float64{50, 40}, [2]float64{x, y})

	x, y = r.ClampPoint(-5, 500)
	require.Equal(t, [2]float64{10, 70}, [2]float64{x, y})

	x, y = r.ClampPoint(200, 0)
	require.Equal(t, [2]float64{110, 20}, [2]float64{x, y})
}

func TestRectClampRect(t *testing.T) {
	bounds := geom.Rect{X: 0, Y: 0, Width: 100, Height: 50}

	inside := geom.Rect{X: 10, Y: 10, Width: 20, Height: 20}
	require.Equal(t, inside, bounds.ClampRect(inside))

	// Pushed back in from each side, size kept.
	require.Equal(t, geom.Rect{X: 0, Y: 0, Width: 20, Height: 20},
		bounds.ClampRect(geom.Rect{X: -15, Y: -5, Width: 20, Height: 20}))
	require.Equal(t, geom.Rect{X: 80, Y: 30, Width: 20, Height: 20},
		bounds.ClampRect(geom.Rect{X: 95, Y: 40, Width: 20, Height: 20}))

	// Wider than bounds: centered on x, still clamped on y.
	require.Equal(t, geom.Rect{X: -10, Y: 30, Width: 120, Height: 20},
		bounds.ClampRect(geom.Rect{X: 40, Y: 45, Width: 120, Height: 20}))

	// Bigger on both axes: centered.
	require.Equal(t, geom.Rect{X: -50, Y: -25, Width: 200, Height: 100},
		bounds.ClampRect(geom.Rect{X: 300, Y: 300, Width: 200, Height: 100}))

	// Same size: snaps onto bounds.
	require.Equal(t, bounds, bounds.ClampRect(geom.Rect{X: 7, Y: -3, Width: 100, Height: 50}))
}
