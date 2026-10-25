package display_test

import (
	"testing"

	"github.com/go-openapi/testify/require"
	"github.com/hajimehoshi/ebiten/v2"

	"github.com/adm87/finch/pkg/display"
	"github.com/adm87/finch/pkg/geom"
)

func requirePoint(t *testing.T, wantX, wantY, gotX, gotY float64) {
	t.Helper()
	require.InDelta(t, wantX, gotX, 1e-9)
	require.InDelta(t, wantY, gotY, 1e-9)
}

func TestScreenFitReturnsTheWindowSize(t *testing.T) {
	screen := display.NewScreen(320, 180)

	w, h := screen.Fit(1280, 720)
	require.Equal(t, 1280.0, w)
	require.Equal(t, 720.0, h)
	require.Equal(t, 4.0, screen.Scale())
	require.Equal(t, geom.Rect{Width: 320, Height: 180}, screen.SafeArea())
	require.Equal(t, ebiten.FilterNearest, screen.Options().Filter)

	x, y := screen.BufferToWindow(10, 5)
	requirePoint(t, 40, 20, x, y)
}

func TestScreenFitCropsToFillTheWindow(t *testing.T) {
	cases := []struct {
		name          string
		width, height float64
		safeArea      geom.Rect
	}{
		{"wider window crops top and bottom", 1280, 600, geom.Rect{X: 0, Y: 15, Width: 320, Height: 150}},
		{"taller window crops the sides", 1280, 800, geom.Rect{X: 16, Y: 0, Width: 288, Height: 180}},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			screen := display.NewScreen(320, 180)
			screen.Fit(tc.width, tc.height)

			got := screen.SafeArea()
			require.InDelta(t, tc.safeArea.X, got.X, 1e-9)
			require.InDelta(t, tc.safeArea.Y, got.Y, 1e-9)
			require.InDelta(t, tc.safeArea.Width, got.Width, 1e-9)
			require.InDelta(t, tc.safeArea.Height, got.Height, 1e-9)

			// The safe area starts at the window corner.
			x, y := screen.BufferToWindow(got.X, got.Y)
			requirePoint(t, 0, 0, x, y)
		})
	}
}

func TestScreenDrawOptionsMatchBufferToWindow(t *testing.T) {
	screen := display.NewScreen(320, 180)
	screen.Fit(1280, 800)

	for _, p := range [][2]float64{{0, 0}, {16, 0}, {100, 90}, {320, 180}} {
		wantX, wantY := screen.BufferToWindow(p[0], p[1])
		gotX, gotY := screen.Options().GeoM.Apply(p[0], p[1])
		requirePoint(t, wantX, wantY, gotX, gotY)

		bx, by := screen.WindowToBuffer(wantX, wantY)
		requirePoint(t, p[0], p[1], bx, by)
	}
}

func TestScreenFitIgnoresAnEmptyWindow(t *testing.T) {
	screen := display.NewScreen(320, 180)
	screen.Fit(1280, 720)

	screen.Fit(0, 0)
	require.Equal(t, 4.0, screen.Scale())
	w, h := screen.WindowSize()
	require.Equal(t, 1280.0, w)
	require.Equal(t, 720.0, h)
}

func TestScreenSetResolutionScaleIgnoresNonPositive(t *testing.T) {
	screen := display.NewScreen(320, 180)
	require.Equal(t, 1.0, screen.ResolutionScale())

	screen.SetResolutionScale(0.5)
	screen.SetResolutionScale(0)
	screen.SetResolutionScale(-1)
	require.Equal(t, 0.5, screen.ResolutionScale())
}
