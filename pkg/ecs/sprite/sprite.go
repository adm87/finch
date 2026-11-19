package sprite

import (
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/yohamta/donburi"

	"github.com/adm87/finch/pkg/assets"
	"github.com/adm87/finch/pkg/images"
)

// Sprite is a texture frame drawn at an entity's transform.
type Sprite struct {
	texture assets.Handle[*images.Texture]
	frame   int
	anchorX float64
	anchorY float64
	color   ebiten.ColorScale
}

var Component = donburi.NewComponentType[Sprite]()

func (s *Sprite) Texture() assets.Handle[*images.Texture] {
	return s.texture
}

func (s *Sprite) Frame() int {
	return s.frame
}

func (s *Sprite) Color() ebiten.ColorScale {
	return s.color
}

// Anchor is the pivot as a fraction of the frame size, so 0.5, 0.5 is the centre.
func (s *Sprite) Anchor() (x, y float64) {
	return s.anchorX, s.anchorY
}

func (s *Sprite) SetTexture(texture assets.Handle[*images.Texture]) {
	s.texture = texture
}

func (s *Sprite) SetFrame(frame int) {
	s.frame = frame
}

func (s *Sprite) SetColor(c ebiten.ColorScale) {
	s.color = c
}

func (s *Sprite) SetAnchor(x, y float64) {
	s.anchorX = x
	s.anchorY = y
}
