package display

import (
	"image/color"
	"math"

	"github.com/hajimehoshi/ebiten/v2"

	"github.com/adm87/finch/pkg/engine"
	"github.com/adm87/finch/pkg/geom"
)

// Screen is the buffer the game draws into. It's scaled to fill the window,
// cropping whatever edges don't fit.
type Screen struct {
	buffer          *ebiten.Image
	options         ebiten.DrawImageOptions
	visible         geom.Rect
	scale           float64
	deviceScale     float64
	resolutionScale float64
	windowWidth     float64
	windowHeight    float64
	dirty           bool
	clearColor      color.RGBA
}

func NewScreen(config ScreenConfig) *Screen {
	if config.ResolutionScale <= 0 {
		panic("display: resolution scale must be positive")
	}
	s := &Screen{
		buffer: ebiten.NewImage(config.Width, config.Height),
		visible: geom.Rect{
			Width:  float64(config.Width),
			Height: float64(config.Height),
		},
		scale:           1,
		deviceScale:     1,
		resolutionScale: config.ResolutionScale,
		dirty:           true,
		clearColor:      color.RGBA(config.ClearColor),
	}
	s.options.Filter = ebiten.FilterNearest
	return s
}

// WindowSize is in window pixels.
func (s *Screen) WindowSize() (float64, float64) {
	return s.windowWidth, s.windowHeight
}

// Visible is the part of the buffer the window shows, rounded out to include
// partly visible pixels.
func (s *Screen) Visible() geom.Rect {
	return s.visible
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

	s.dirty = true
}

// ResolutionScale is the window's resolution relative to the display. 1 draws
// at full resolution and 0.5 at half.
func (s *Screen) ResolutionScale() float64 {
	return s.resolutionScale
}

func (s *Screen) Filter() ebiten.Filter {
	return s.options.Filter
}

func (s *Screen) SetFilter(filter ebiten.Filter) {
	s.options.Filter = filter
}

// Layout takes the window size from Ebiten and returns it in window pixels.
func (s *Screen) Layout(outsideWidth, outsideHeight float64) (float64, float64) {
	deviceScale := s.deviceScale
	if m := ebiten.Monitor(); m != nil {
		deviceScale = m.DeviceScaleFactor()
	}
	return s.layout(outsideWidth, outsideHeight, deviceScale)
}

// CursorPosition is the mouse position in window pixels.
func (s *Screen) CursorPosition() (float64, float64) {
	x, y := ebiten.CursorPosition()
	return float64(x), float64(y)
}

func (s *Screen) BufferToWindow(x, y float64) (float64, float64) {
	return (x - s.visible.X) * s.scale, (y - s.visible.Y) * s.scale
}

func (s *Screen) WindowToBuffer(x, y float64) (float64, float64) {
	return x/s.scale + s.visible.X, y/s.scale + s.visible.Y
}

// Present draws the passes before it into the buffer, then draws the buffer into the target.
func (s *Screen) Present() engine.DrawPass {
	return func(inner engine.Drawer) engine.Drawer {
		return engine.DrawFunc(func(target *ebiten.Image) {
			s.buffer.Fill(s.clearColor)
			inner.Draw(s.buffer)
			target.DrawImage(s.buffer, &s.options)
		})
	}
}

// layout is Layout without the monitor, so tests can pick the device scale.
func (s *Screen) layout(outsideWidth, outsideHeight, deviceScale float64) (float64, float64) {
	s.deviceScale = deviceScale
	scale := s.pixelScale()
	return s.fit(outsideWidth*scale, outsideHeight*scale)
}

func (s *Screen) fit(windowWidth, windowHeight float64) (float64, float64) {
	if windowWidth <= 0 || windowHeight <= 0 {
		return windowWidth, windowHeight
	}
	if s.dirty || s.windowWidth != windowWidth || s.windowHeight != windowHeight {
		s.calculateLayout(windowWidth, windowHeight)
		s.windowWidth, s.windowHeight = windowWidth, windowHeight
		s.dirty = false
	}
	return windowWidth, windowHeight
}

// pixelScale is window pixels per logical pixel.
func (s *Screen) pixelScale() float64 {
	return s.deviceScale * s.resolutionScale
}

func (s *Screen) calculateLayout(windowWidth, windowHeight float64) {
	buffW, buffH := float64(s.buffer.Bounds().Dx()), float64(s.buffer.Bounds().Dy())

	s.scale = max(windowWidth/buffW, windowHeight/buffH)

	visibleW, visibleH := windowWidth/s.scale, windowHeight/s.scale
	s.visible.X = math.Floor((buffW - visibleW) / 2)
	s.visible.Y = math.Floor((buffH - visibleH) / 2)
	s.visible.Width = math.Ceil(s.visible.X+visibleW) - s.visible.X
	s.visible.Height = math.Ceil(s.visible.Y+visibleH) - s.visible.Y

	s.options.GeoM.Reset()
	s.options.GeoM.Translate(-s.visible.X, -s.visible.Y)
	s.options.GeoM.Scale(s.scale, s.scale)
}
