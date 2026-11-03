package assets

import (
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
)

var missingColor = color.RGBA{R: 128, A: 255, B: 128}

// Images is a placeholder image store until real assets load.
type Images struct {
	missing *ebiten.Image
}

func NewImages() *Images {
	missing := ebiten.NewImage(10, 10)
	missing.Fill(missingColor)
	return &Images{missing: missing}
}

// Image returns the missing texture for every handle.
func (i *Images) Image(_ uint64) *ebiten.Image {
	return i.missing
}
