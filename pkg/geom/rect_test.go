package geom_test

import (
	"testing"

	"github.com/adm87/finch/pkg/geom"
)

func TestRectCorners(t *testing.T) {
	r := geom.Rect{X: 10, Y: 20, Width: 30, Height: 40}

	if x, y := r.Min(); x != 10 || y != 20 {
		t.Errorf("Min() = (%v, %v), want (10, 20)", x, y)
	}
	if x, y := r.Max(); x != 40 || y != 60 {
		t.Errorf("Max() = (%v, %v), want (40, 60)", x, y)
	}
}

func TestRectIntersects(t *testing.T) {
	r := geom.Rect{X: 0, Y: 0, Width: 10, Height: 10}

	tests := []struct {
		name string
		o    geom.Rect
		want bool
	}{
		{"overlapping", geom.Rect{X: 5, Y: 5, Width: 10, Height: 10}, true},
		{"inside", geom.Rect{X: 2, Y: 2, Width: 2, Height: 2}, true},
		{"around", geom.Rect{X: -5, Y: -5, Width: 20, Height: 20}, true},
		{"touching edge", geom.Rect{X: 10, Y: 0, Width: 5, Height: 5}, false},
		{"apart", geom.Rect{X: 20, Y: 20, Width: 5, Height: 5}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := r.Intersects(tt.o); got != tt.want {
				t.Errorf("Intersects(%+v) = %v, want %v", tt.o, got, tt.want)
			}
			if got := tt.o.Intersects(r); got != tt.want {
				t.Errorf("reversed Intersects = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRectCenterSizeArea(t *testing.T) {
	r := geom.Rect{X: 10, Y: 20, Width: 30, Height: 40}

	if x, y := r.Center(); x != 25 || y != 40 {
		t.Errorf("Center() = (%v, %v), want (25, 40)", x, y)
	}
	if w, h := r.Size(); w != 30 || h != 40 {
		t.Errorf("Size() = (%v, %v), want (30, 40)", w, h)
	}
	if got := r.Area(); got != 1200 {
		t.Errorf("Area() = %v, want 1200", got)
	}
}

func TestRectContains(t *testing.T) {
	r := geom.Rect{X: 0, Y: 0, Width: 10, Height: 10}

	tests := []struct {
		name string
		x, y float64
		want bool
	}{
		{"inside", 5, 5, true},
		{"min corner", 0, 0, true},
		{"max edge", 10, 5, false},
		{"max corner", 10, 10, false},
		{"left", -1, 5, false},
		{"below", 5, 11, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := r.Contains(tt.x, tt.y); got != tt.want {
				t.Errorf("Contains(%v, %v) = %v, want %v", tt.x, tt.y, got, tt.want)
			}
		})
	}
}
