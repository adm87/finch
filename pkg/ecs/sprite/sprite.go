package sprite

import (
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/yohamta/donburi"

	"github.com/adm87/finch/pkg/assets"
)

// Sprite is an image drawn at an entity's transform.
type Sprite struct {
	handle  assets.Handle[*ebiten.Image]
	anchorX float64
	anchorY float64
	color   ebiten.ColorScale
}

var Component = donburi.NewComponentType[Sprite]()

func (s *Sprite) Handle() assets.Handle[*ebiten.Image] {
	return s.handle
}

func (s *Sprite) Color() ebiten.ColorScale {
	return s.color
}

// Anchor is the pivot as a fraction of the image size, so 0.5, 0.5 is the centre.
func (s *Sprite) Anchor() (x, y float64) {
	return s.anchorX, s.anchorY
}

func (s *Sprite) SetHandle(handle assets.Handle[*ebiten.Image]) {
	s.handle = handle
}

func (s *Sprite) SetColor(c ebiten.ColorScale) {
	s.color = c
}

func (s *Sprite) SetAnchor(x, y float64) {
	s.anchorX = x
	s.anchorY = y
}
