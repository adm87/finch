package display

import (
	"math"

	"github.com/hajimehoshi/ebiten/v2"

	"github.com/adm87/finch/pkg/geom"
)

type Screen struct {
	buffer    *ebiten.Image
	options   *ebiten.DrawImageOptions
	safeArea  geom.Rect
	isDirty   bool
	oldWidth  int
	oldHeight int
}

func NewScreen(width, height int) *Screen {
	return &Screen{
		buffer:  ebiten.NewImage(width, height),
		options: &ebiten.DrawImageOptions{},
		safeArea: geom.Rect{
			X:      0,
			Y:      0,
			Width:  float64(width),
			Height: float64(height),
		},
		isDirty: true,
	}
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

func (s *Screen) Resize(width, height int) {
	buffW, buffH := s.buffer.Bounds().Dx(), s.buffer.Bounds().Dy()
	if width == buffW && height == buffH {
		return
	}

	s.buffer.Deallocate()
	s.buffer = ebiten.NewImage(width, height)

	s.isDirty = true
}

func (s *Screen) Layout(outsideWidth, outsideHeight int) (int, int) {
	if s.isDirty || s.oldWidth != outsideWidth || s.oldHeight != outsideHeight {
		s.calculateLayout(outsideWidth, outsideHeight)
		s.oldWidth, s.oldHeight = outsideWidth, outsideHeight
		s.isDirty = false
	}
	return int(s.safeArea.Width), int(s.safeArea.Height)
}

func (s *Screen) ScreenToSafeArea(x, y float64) (float64, float64) {
	return x + s.safeArea.X, y + s.safeArea.Y
}

func (s *Screen) SafeAreaToScreen(x, y float64) (float64, float64) {
	return x - s.safeArea.X, y - s.safeArea.Y
}

func (s *Screen) calculateLayout(outsideWidth, outsideHeight int) {
	buffW, buffH := s.buffer.Bounds().Dx(), s.buffer.Bounds().Dy()

	scale := max(
		float64(outsideWidth)/float64(buffW),
		float64(outsideHeight)/float64(buffH),
	)

	s.safeArea.Width = float64(outsideWidth) / scale
	s.safeArea.Height = float64(outsideHeight) / scale

	s.safeArea.X = math.Floor((float64(buffW) - float64(s.safeArea.Width)) / 2)
	s.safeArea.Y = math.Floor((float64(buffH) - float64(s.safeArea.Height)) / 2)

	s.options.GeoM.Reset()
	s.options.GeoM.Translate(-s.safeArea.X, -s.safeArea.Y)
}
