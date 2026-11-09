package sprite

import (
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/yohamta/donburi"
)

// Sprite is an image drawn at an entity's transform.
type Sprite struct {
	handle  uint64
	anchorX float64
	anchorY float64
	color   ebiten.ColorScale
}

var Component = donburi.NewComponentType[Sprite]()

func (s *Sprite) Handle() uint64 {
	return s.handle
}

func (s *Sprite) Color() ebiten.ColorScale {
	return s.color
}

// Anchor is the pivot as a fraction of the image size, so 0.5, 0.5 is the centre.
func (s *Sprite) Anchor() (x, y float64) {
	return s.anchorX, s.anchorY
}

func (s *Sprite) SetHandle(handle uint64) {
	s.handle = handle
}

func (s *Sprite) SetColor(c ebiten.ColorScale) {
	s.color = c
}

func (s *Sprite) SetAnchor(x, y float64) {
	s.anchorX = x
	s.anchorY = y
}
