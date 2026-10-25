package display

import (
	"math"

	"github.com/hajimehoshi/ebiten/v2"

	"github.com/adm87/finch/pkg/geom"
)

// Screen is the buffer the game draws into. It's scaled to fill the window,
// cropping whatever edges don't fit.
type Screen struct {
	buffer  *ebiten.Image
	options *ebiten.DrawImageOptions
	// safeArea is the visible part of the buffer.
	safeArea     geom.Rect
	scale        float64
	deviceScale  float64
	windowWidth  float64
	windowHeight float64
	isDirty      bool
}

func NewScreen(width, height int) *Screen {
	options := &ebiten.DrawImageOptions{}
	options.Filter = ebiten.FilterNearest

	return &Screen{
		buffer:  ebiten.NewImage(width, height),
		options: options,
		safeArea: geom.Rect{
			Width:  float64(width),
			Height: float64(height),
		},
		scale:       1,
		deviceScale: 1,
		isDirty:     true,
	}
}

// WindowSize is in device pixels.
func (s *Screen) WindowSize() (float64, float64) {
	return s.windowWidth, s.windowHeight
}

func (s *Screen) Buffer() *ebiten.Image {
	return s.buffer
}

func (s *Screen) Options() *ebiten.DrawImageOptions {
	return s.options
}

func (s *Screen) SafeArea() geom.Rect {
	return s.safeArea
}

// Scale is window pixels per buffer pixel.
func (s *Screen) Scale() float64 {
	return s.scale
}

func (s *Screen) Resize(width, height int) {
	buffW, buffH := s.buffer.Bounds().Dx(), s.buffer.Bounds().Dy()
	if width == buffW && height == buffH {
		return
	}

	s.buffer.Deallocate()
	s.buffer = ebiten.NewImage(width, height)

	s.isDirty = true
}

// Layout takes the window size from Ebiten and returns it in device pixels, so
// the window is drawn at full resolution.
func (s *Screen) Layout(outsideWidth, outsideHeight float64) (float64, float64) {
	if m := ebiten.Monitor(); m != nil {
		s.deviceScale = m.DeviceScaleFactor()
	}
	return s.Fit(outsideWidth*s.deviceScale, outsideHeight*s.deviceScale)
}

// Fit scales the buffer to a window size in device pixels.
func (s *Screen) Fit(windowWidth, windowHeight float64) (float64, float64) {
	if windowWidth <= 0 || windowHeight <= 0 {
		return windowWidth, windowHeight
	}
	if s.isDirty || s.windowWidth != windowWidth || s.windowHeight != windowHeight {
		s.calculateLayout(windowWidth, windowHeight)
		s.windowWidth, s.windowHeight = windowWidth, windowHeight
		s.isDirty = false
	}
	return windowWidth, windowHeight
}

// CursorPosition is the mouse position in device pixels.
func (s *Screen) CursorPosition() (float64, float64) {
	x, y := ebiten.CursorPosition()
	return float64(x) * s.deviceScale, float64(y) * s.deviceScale
}

func (s *Screen) BufferToWindow(x, y float64) (float64, float64) {
	return (x - s.safeArea.X) * s.scale, (y - s.safeArea.Y) * s.scale
}

func (s *Screen) WindowToBuffer(x, y float64) (float64, float64) {
	return x/s.scale + s.safeArea.X, y/s.scale + s.safeArea.Y
}

func (s *Screen) calculateLayout(windowWidth, windowHeight float64) {
	buffW, buffH := float64(s.buffer.Bounds().Dx()), float64(s.buffer.Bounds().Dy())

	s.scale = max(windowWidth/buffW, windowHeight/buffH)

	s.safeArea.Width = windowWidth / s.scale
	s.safeArea.Height = windowHeight / s.scale
	s.safeArea.X = math.Floor((buffW - s.safeArea.Width) / 2)
	s.safeArea.Y = math.Floor((buffH - s.safeArea.Height) / 2)

	s.options.GeoM.Reset()
	s.options.GeoM.Translate(-s.safeArea.X, -s.safeArea.Y)
	s.options.GeoM.Scale(s.scale, s.scale)
}
