package bounds_test

import (
	"math"
	"testing"

	"github.com/hajimehoshi/ebiten/v2"

	"github.com/adm87/finch/pkg/ecs/bounds"
	"github.com/adm87/finch/pkg/geom"
)

func TestWorld(t *testing.T) {
	local := geom.Rect{X: -8, Y: -8, Width: 16, Height: 16}

	tests := []struct {
		name string
		m    func(*ebiten.GeoM)
		want geom.Rect
	}{
		{"identity", func(*ebiten.GeoM) {}, local},
		{"translate", func(m *ebiten.GeoM) { m.Translate(100, 50) }, geom.Rect{X: 92, Y: 42, Width: 16, Height: 16}},
		{"scale", func(m *ebiten.GeoM) { m.Scale(2, 3) }, geom.Rect{X: -16, Y: -24, Width: 32, Height: 48}},
		{"flip", func(m *ebiten.GeoM) { m.Scale(-1, -1) }, local},
		{"rotate 45", func(m *ebiten.GeoM) { m.Rotate(math.Pi / 4) }, geom.Rect{
			X: -8 * math.Sqrt2, Y: -8 * math.Sqrt2, Width: 16 * math.Sqrt2, Height: 16 * math.Sqrt2,
		}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var b bounds.Bounds
			b.SetLocal(local)

			var m ebiten.GeoM
			tt.m(&m)

			if got := b.World(m); !nearRect(got, tt.want) {
				t.Errorf("World() = %+v, want %+v", got, tt.want)
			}
		})
	}
}

func nearRect(a, b geom.Rect) bool {
	return near(a.X, b.X) && near(a.Y, b.Y) && near(a.Width, b.Width) && near(a.Height, b.Height)
}

func near(a, b float64) bool {
	return math.Abs(a-b) < 1e-9
}
