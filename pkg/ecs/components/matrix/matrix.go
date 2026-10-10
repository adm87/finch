package matrix

import (
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/yohamta/donburi"
)

type Matrix struct {
	m ebiten.GeoM
}

var ComponentType = donburi.NewComponentType[Matrix]()

func (m *Matrix) GeoM() ebiten.GeoM {
	return m.m
}

func (m *Matrix) Update(x, y, sx, sy, ox, oy, rotation float64) {
	m.m.Reset()
	m.m.Translate(-ox, -oy)
	m.m.Scale(sx, sy)
	m.m.Rotate(rotation)
	m.m.Translate(x, y)
}
