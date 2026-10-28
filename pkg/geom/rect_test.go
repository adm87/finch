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
