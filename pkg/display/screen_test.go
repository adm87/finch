package display

import (
	"math"
	"testing"

	"github.com/hajimehoshi/ebiten/v2"

	"github.com/adm87/finch/pkg/engine"
	"github.com/adm87/finch/pkg/geom"
)

func TestNewScreenDrawsNearest(t *testing.T) {
	s := NewScreen(ScreenConfig{Width: 320, Height: 180, ResolutionScale: 1})
	if got := s.Filter(); got != ebiten.FilterNearest {
		t.Errorf("Filter() = %v, want FilterNearest", got)
	}
}

func TestScreenFitScalesToTheWindow(t *testing.T) {
	s := NewScreen(ScreenConfig{Width: 320, Height: 180, ResolutionScale: 1})

	if w, h := s.fit(1280, 720); w != 1280 || h != 720 {
		t.Errorf("fit() = (%v, %v), want (1280, 720)", w, h)
	}
	if got := s.Scale(); got != 4 {
		t.Errorf("Scale() = %v, want 4", got)
	}
	if got, want := s.Visible(), (geom.Rect{Width: 320, Height: 180}); got != want {
		t.Errorf("Visible() = %+v, want %+v", got, want)
	}
	if x, y := s.BufferToWindow(10, 5); !near(x, 40) || !near(y, 20) {
		t.Errorf("BufferToWindow(10, 5) = (%v, %v), want (40, 20)", x, y)
	}
}

func TestScreenFitCropsToFillTheWindow(t *testing.T) {
	tests := []struct {
		name          string
		width, height float64
		want          geom.Rect
	}{
		{"wider window crops top and bottom", 1280, 600, geom.Rect{X: 0, Y: 15, Width: 320, Height: 150}},
		{"taller window crops the sides", 1280, 800, geom.Rect{X: 16, Y: 0, Width: 288, Height: 180}},
		{"fractional crop rounds out to whole pixels", 1300, 720, geom.Rect{X: 0, Y: 1, Width: 320, Height: 178}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := NewScreen(ScreenConfig{Width: 320, Height: 180, ResolutionScale: 1})
			s.fit(tt.width, tt.height)

			got := s.Visible()
			if !nearRect(got, tt.want) {
				t.Errorf("Visible() = %+v, want %+v", got, tt.want)
			}
			// The visible area starts at the window corner.
			if x, y := s.BufferToWindow(got.X, got.Y); !near(x, 0) || !near(y, 0) {
				t.Errorf("BufferToWindow(visible corner) = (%v, %v), want (0, 0)", x, y)
			}
		})
	}
}

func TestScreenLayoutScalesByDeviceAndResolution(t *testing.T) {
	tests := []struct {
		name                         string
		deviceScale, resolutionScale float64
		wantW, wantH, wantScale      float64
	}{
		{"standard display", 1, 1, 640, 360, 2},
		{"high-DPI display", 2, 1, 1280, 720, 4},
		{"high-DPI at half resolution", 2, 0.5, 640, 360, 2},
		{"standard display at half resolution", 1, 0.5, 320, 180, 1},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := NewScreen(ScreenConfig{Width: 320, Height: 180, ResolutionScale: tt.resolutionScale})

			if w, h := s.layout(640, 360, tt.deviceScale); w != tt.wantW || h != tt.wantH {
				t.Errorf("layout() = (%v, %v), want (%v, %v)", w, h, tt.wantW, tt.wantH)
			}
			if got := s.Scale(); got != tt.wantScale {
				t.Errorf("Scale() = %v, want %v", got, tt.wantScale)
			}
			if w, h := s.WindowSize(); w != tt.wantW || h != tt.wantH {
				t.Errorf("WindowSize() = (%v, %v), want (%v, %v)", w, h, tt.wantW, tt.wantH)
			}
		})
	}
}

func TestScreenLayoutRefitsWhenDeviceScaleChanges(t *testing.T) {
	s := NewScreen(ScreenConfig{Width: 320, Height: 180, ResolutionScale: 1})
	s.layout(640, 360, 1)

	// Same window, but dragged onto a high-DPI monitor.
	s.layout(640, 360, 2)

	if got := s.Scale(); got != 4 {
		t.Errorf("Scale() = %v, want 4", got)
	}
}

func TestScreenTransformMatchesBufferToWindow(t *testing.T) {
	s := NewScreen(ScreenConfig{Width: 320, Height: 180, ResolutionScale: 1})
	s.fit(1280, 800)

	for _, p := range [][2]float64{{0, 0}, {16, 0}, {100, 90}, {320, 180}} {
		wantX, wantY := s.BufferToWindow(p[0], p[1])
		if x, y := s.options.GeoM.Apply(p[0], p[1]); !near(x, wantX) || !near(y, wantY) {
			t.Errorf("GeoM.Apply(%v) = (%v, %v), want (%v, %v)", p, x, y, wantX, wantY)
		}
		if x, y := s.WindowToBuffer(wantX, wantY); !near(x, p[0]) || !near(y, p[1]) {
			t.Errorf("WindowToBuffer round trip of %v = (%v, %v)", p, x, y)
		}
	}
}

func TestScreenFitIgnoresAnEmptyWindow(t *testing.T) {
	s := NewScreen(ScreenConfig{Width: 320, Height: 180, ResolutionScale: 1})
	s.fit(1280, 720)
	s.fit(0, 0)

	if got := s.Scale(); got != 4 {
		t.Errorf("Scale() = %v, want 4", got)
	}
	if w, h := s.WindowSize(); w != 1280 || h != 720 {
		t.Errorf("WindowSize() = (%v, %v), want (1280, 720)", w, h)
	}
}

func TestScreenResizeRefitsTheSameWindow(t *testing.T) {
	s := NewScreen(ScreenConfig{Width: 320, Height: 180, ResolutionScale: 1})
	s.fit(1280, 720)

	s.Resize(640, 360)
	s.fit(1280, 720)

	if got := s.Scale(); got != 2 {
		t.Errorf("Scale() = %v, want 2", got)
	}
	if got, want := s.Visible(), (geom.Rect{Width: 640, Height: 360}); got != want {
		t.Errorf("Visible() = %+v, want %+v", got, want)
	}
}

func TestScreenResizeToTheSameSizeKeepsTheBuffer(t *testing.T) {
	s := NewScreen(ScreenConfig{Width: 320, Height: 180, ResolutionScale: 1})
	buffer := s.buffer

	s.Resize(320, 180)

	if s.buffer != buffer {
		t.Error("Resize to the same size replaced the buffer")
	}
}

func TestNewScreenRejectsNonPositiveResolutionScale(t *testing.T) {
	for _, scale := range []float64{0, -1} {
		func() {
			defer func() {
				if recover() == nil {
					t.Errorf("NewScreen with resolution scale %v did not panic", scale)
				}
			}()
			NewScreen(ScreenConfig{Width: 320, Height: 180, ResolutionScale: scale})
		}()
	}
}

func TestScreenPresentDrawsIntoTheBuffer(t *testing.T) {
	s := NewScreen(ScreenConfig{Width: 320, Height: 180, ResolutionScale: 1})

	var got *ebiten.Image
	inner := engine.DrawFunc(func(target *ebiten.Image) { got = target })

	s.Present()(inner).Draw(ebiten.NewImage(1280, 720))

	if got != s.buffer {
		t.Error("passes before Present did not draw into the buffer")
	}
}

func near(a, b float64) bool {
	return math.Abs(a-b) < 1e-9
}

func nearRect(a, b geom.Rect) bool {
	return near(a.X, b.X) && near(a.Y, b.Y) && near(a.Width, b.Width) && near(a.Height, b.Height)
}
