package transform_test

import (
	"math"
	"testing"

	"github.com/go-openapi/testify/require"
	"github.com/yohamta/donburi"

	"github.com/adm87/finch/pkg/geom"
	"github.com/adm87/finch/pkg/transform"
)

func newEntity(t *testing.T) (*transform.Transform, *transform.Bounds) {
	t.Helper()
	w := donburi.NewWorld()
	entry := w.Entry(w.Create(transform.TransformComponent, transform.BoundsComponent))

	tr, ok := transform.GetTransform(entry)
	require.True(t, ok)
	b, ok := transform.GetBounds(entry)
	require.True(t, ok)
	return tr, b
}

func requirePoint(t *testing.T, wantX, wantY, gotX, gotY float64) {
	t.Helper()
	require.InDelta(t, wantX, gotX, 1e-9)
	require.InDelta(t, wantY, gotY, 1e-9)
}

func TestTransformDefaultsToIdentity(t *testing.T) {
	tr, _ := newEntity(t)

	sx, sy := tr.Scale()
	require.Equal(t, 1.0, sx)
	require.Equal(t, 1.0, sy)

	geoM := tr.GeoM()
	x, y := geoM.Apply(3, 4)
	requirePoint(t, 3, 4, x, y)
}

func TestTransformAppliesOriginScaleRotationThenPosition(t *testing.T) {
	tr, _ := newEntity(t)
	tr.SetOrigin(10, 0)
	tr.SetScale(2, 2)
	tr.SetRotation(math.Pi / 2)
	tr.SetPosition(100, 50)

	geoM := tr.GeoM()

	x, y := geoM.Apply(10, 0)
	requirePoint(t, 100, 50, x, y)

	// One unit right of the origin, doubled and turned a quarter.
	x, y = geoM.Apply(11, 0)
	requirePoint(t, 100, 52, x, y)
}

func TestTransformGeoMUpdatesAfterChange(t *testing.T) {
	tr, _ := newEntity(t)

	geoM := tr.GeoM()
	x, y := geoM.Apply(0, 0)
	requirePoint(t, 0, 0, x, y)

	tr.SetPosition(5, 6)
	geoM = tr.GeoM()
	x, y = geoM.Apply(0, 0)
	requirePoint(t, 5, 6, x, y)
}

func TestBoundsWorldRect(t *testing.T) {
	cases := []struct {
		name     string
		position [2]float64
		scale    [2]float64
		rotation float64
		want     geom.Rect
	}{
		{"identity", [2]float64{0, 0}, [2]float64{1, 1}, 0, geom.Rect{X: -5, Y: -10, Width: 10, Height: 20}},
		{"translated", [2]float64{7, 3}, [2]float64{1, 1}, 0, geom.Rect{X: 2, Y: -7, Width: 10, Height: 20}},
		{"scaled", [2]float64{0, 0}, [2]float64{2, 3}, 0, geom.Rect{X: -10, Y: -30, Width: 20, Height: 60}},
		{"flipped", [2]float64{0, 0}, [2]float64{-1, 1}, 0, geom.Rect{X: -5, Y: -10, Width: 10, Height: 20}},
		{"quarter turn swaps width and height", [2]float64{0, 0}, [2]float64{1, 1}, math.Pi / 2, geom.Rect{X: -10, Y: -5, Width: 20, Height: 10}},
		{"eighth turn grows to fit", [2]float64{0, 0}, [2]float64{1, 1}, math.Pi / 4, geom.Rect{
			X: -15 / math.Sqrt2, Y: -15 / math.Sqrt2, Width: 30 / math.Sqrt2, Height: 30 / math.Sqrt2,
		}},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			tr, b := newEntity(t)
			b.SetRect(geom.Rect{X: -5, Y: -10, Width: 10, Height: 20})
			tr.SetPosition(tc.position[0], tc.position[1])
			tr.SetScale(tc.scale[0], tc.scale[1])
			tr.SetRotation(tc.rotation)

			got := b.WorldRect(tr)
			require.InDelta(t, tc.want.X, got.X, 1e-9)
			require.InDelta(t, tc.want.Y, got.Y, 1e-9)
			require.InDelta(t, tc.want.Width, got.Width, 1e-9)
			require.InDelta(t, tc.want.Height, got.Height, 1e-9)
		})
	}
}

func TestVersionsChangeOnlyWhenValuesChange(t *testing.T) {
	tr, b := newEntity(t)

	v := tr.Version()
	tr.SetPosition(0, 0)
	tr.SetScale(1, 1)
	tr.SetRotation(0)
	tr.SetOrigin(0, 0)
	require.Equal(t, v, tr.Version())

	tr.SetPosition(1, 0)
	require.Greater(t, tr.Version(), v)
	v = tr.Version()
	tr.SetScale(-1, 1)
	tr.SetRotation(1)
	tr.SetOrigin(2, 2)
	require.Equal(t, v+3, tr.Version())

	bv := b.Version()
	b.SetRect(b.Rect())
	require.Equal(t, bv, b.Version())
	b.SetRect(geom.Rect{Width: 5, Height: 5})
	require.Greater(t, b.Version(), bv)
}

func TestGeoMUpdatesAfterEachSetter(t *testing.T) {
	tr, _ := newEntity(t)
	tr.SetPosition(10, 0)

	apply := func(x, y float64) (float64, float64) {
		geoM := tr.GeoM()
		return geoM.Apply(x, y)
	}

	x, y := apply(0, 0)
	requirePoint(t, 10, 0, x, y)

	tr.SetPosition(20, 5)
	x, y = apply(0, 0)
	requirePoint(t, 20, 5, x, y)

	tr.SetScale(2, 3)
	x, y = apply(1, 1)
	requirePoint(t, 22, 8, x, y)

	tr.SetOrigin(1, 1)
	x, y = apply(1, 1)
	requirePoint(t, 20, 5, x, y)

	tr.SetRotation(math.Pi)
	x, y = apply(2, 1)
	requirePoint(t, 18, 5, x, y)

	first := tr.GeoM()
	require.Equal(t, first, tr.GeoM())
}
