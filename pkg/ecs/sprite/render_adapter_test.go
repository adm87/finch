package sprite

import (
	"math"
	"testing"

	"github.com/hajimehoshi/ebiten/v2"

	"github.com/adm87/finch/pkg/geom"
)

func TestBounds(t *testing.T) {
	tests := []struct {
		name string
		geo  func(*ebiten.GeoM)
		want geom.Rect
	}{
		{"unmoved", func(*ebiten.GeoM) {}, geom.Rect{Width: 4, Height: 2}},
		{"moved", func(g *ebiten.GeoM) { g.Translate(10, 20) }, geom.Rect{X: 10, Y: 20, Width: 4, Height: 2}},
		{"scaled", func(g *ebiten.GeoM) { g.Scale(2, 3) }, geom.Rect{Width: 8, Height: 6}},
		{"rotated a quarter turn", func(g *ebiten.GeoM) { g.Rotate(math.Pi / 2) }, geom.Rect{X: -2, Width: 2, Height: 4}},
		{"rotated an eighth turn", func(g *ebiten.GeoM) { g.Rotate(math.Pi / 4) }, geom.Rect{
			X: -math.Sqrt2, Width: 3 * math.Sqrt2, Height: 3 * math.Sqrt2,
		}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var geo ebiten.GeoM
			tt.geo(&geo)

			if got := bounds(&geo, 4, 2); !nearRect(got, tt.want) {
				t.Errorf("bounds() = %+v, want %+v", got, tt.want)
			}
		})
	}
}

func nearRect(a, b geom.Rect) bool {
	near := func(x, y float64) bool { return math.Abs(x-y) < 1e-9 }
	return near(a.X, b.X) && near(a.Y, b.Y) && near(a.Width, b.Width) && near(a.Height, b.Height)
}
