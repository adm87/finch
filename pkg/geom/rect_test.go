package geom_test

import (
	"image"
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
	if x, y := r.Center(); x != 25 || y != 40 {
		t.Errorf("Center() = (%v, %v), want (25, 40)", x, y)
	}
	if got := r.Area(); got != 1200 {
		t.Errorf("Area() = %v, want 1200", got)
	}
}

func TestRectEmpty(t *testing.T) {
	tests := []struct {
		r    geom.Rect
		want bool
	}{
		{geom.Rect{}, true},
		{geom.Rect{Width: 5}, true},
		{geom.Rect{Width: 5, Height: -1}, true},
		{geom.Rect{Width: 1, Height: 1}, false},
	}
	for _, tt := range tests {
		if got := tt.r.Empty(); got != tt.want {
			t.Errorf("%+v.Empty() = %v, want %v", tt.r, got, tt.want)
		}
	}
}

func TestRectUnion(t *testing.T) {
	a := geom.Rect{X: 0, Y: 0, Width: 10, Height: 10}
	b := geom.Rect{X: 5, Y: -5, Width: 10, Height: 10}
	far := geom.Rect{X: 32, Y: 32, Width: 8, Height: 8}

	tests := []struct {
		name string
		r, o geom.Rect
		want geom.Rect
	}{
		{"overlapping", a, b, geom.Rect{X: 0, Y: -5, Width: 15, Height: 15}},
		{"swapped", b, a, geom.Rect{X: 0, Y: -5, Width: 15, Height: 15}},
		{"contained", a, geom.Rect{X: 2, Y: 2, Width: 1, Height: 1}, a},
		{"zero starts a union", geom.Rect{}, far, far},
		{"zero adds nothing", far, geom.Rect{}, far},
		{"empty adds nothing wherever it is", far, geom.Rect{X: -100, Y: -100}, far},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.r.Union(tt.o); got != tt.want {
				t.Errorf("Union() = %+v, want %+v", got, tt.want)
			}
		})
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

func TestRectContains(t *testing.T) {
	r := geom.Rect{X: 0, Y: 0, Width: 10, Height: 10}

	tests := []struct {
		name string
		o    geom.Rect
		want bool
	}{
		{"inside", geom.Rect{X: 2, Y: 2, Width: 2, Height: 2}, true},
		{"itself", r, true},
		{"overlapping", geom.Rect{X: 5, Y: 5, Width: 10, Height: 10}, false},
		{"around", geom.Rect{X: -5, Y: -5, Width: 20, Height: 20}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := r.Contains(tt.o); got != tt.want {
				t.Errorf("Contains(%+v) = %v, want %v", tt.o, got, tt.want)
			}
		})
	}
}

func TestRectContainsPoint(t *testing.T) {
	r := geom.Rect{X: 0, Y: 0, Width: 10, Height: 10}

	tests := []struct {
		x, y float64
		want bool
	}{
		{5, 5, true},
		{0, 0, true},
		{-1, 5, false},
		{5, 11, false},
	}
	for _, tt := range tests {
		if got := r.ContainsPoint(tt.x, tt.y); got != tt.want {
			t.Errorf("ContainsPoint(%v, %v) = %v, want %v", tt.x, tt.y, got, tt.want)
		}
	}
}

func TestRectClampPoint(t *testing.T) {
	r := geom.Rect{X: 10, Y: 20, Width: 100, Height: 50}

	tests := []struct {
		x, y         float64
		wantX, wantY float64
	}{
		{50, 40, 50, 40},
		{-5, 500, 10, 70},
		{200, 0, 110, 20},
	}
	for _, tt := range tests {
		x, y := r.ClampPoint(tt.x, tt.y)
		if x != tt.wantX || y != tt.wantY {
			t.Errorf("ClampPoint(%v, %v) = (%v, %v), want (%v, %v)", tt.x, tt.y, x, y, tt.wantX, tt.wantY)
		}
	}
}

func TestRectClampRect(t *testing.T) {
	bounds := geom.Rect{X: 0, Y: 0, Width: 100, Height: 50}

	tests := []struct {
		name string
		o    geom.Rect
		want geom.Rect
	}{
		{"inside is unchanged",
			geom.Rect{X: 10, Y: 10, Width: 20, Height: 20},
			geom.Rect{X: 10, Y: 10, Width: 20, Height: 20}},
		{"pushed in from the top left",
			geom.Rect{X: -15, Y: -5, Width: 20, Height: 20},
			geom.Rect{X: 0, Y: 0, Width: 20, Height: 20}},
		{"pushed in from the bottom right",
			geom.Rect{X: 95, Y: 40, Width: 20, Height: 20},
			geom.Rect{X: 80, Y: 30, Width: 20, Height: 20}},
		{"wider than bounds centers on x",
			geom.Rect{X: 40, Y: 45, Width: 120, Height: 20},
			geom.Rect{X: -10, Y: 30, Width: 120, Height: 20}},
		{"bigger than bounds centers",
			geom.Rect{X: 300, Y: 300, Width: 200, Height: 100},
			geom.Rect{X: -50, Y: -25, Width: 200, Height: 100}},
		{"same size snaps onto bounds",
			geom.Rect{X: 7, Y: -3, Width: 100, Height: 50},
			bounds},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := bounds.ClampRect(tt.o); got != tt.want {
				t.Errorf("ClampRect(%+v) = %+v, want %+v", tt.o, got, tt.want)
			}
		})
	}
}

func TestRectToImageRectangle(t *testing.T) {
	r := geom.Rect{X: 1, Y: 2, Width: 3, Height: 4}
	if got, want := r.ToImageRectangle(), image.Rect(1, 2, 4, 6); got != want {
		t.Errorf("ToImageRectangle() = %v, want %v", got, want)
	}
}
