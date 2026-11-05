// Package transform gives entities a position, scale, origin and rotation.
package transform

import (
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/yohamta/donburi"
)

// Transform places an entity in the world. Its matrix is rebuilt by Update.
type Transform struct {
	x, y          float64
	sx, sy        float64
	ox, oy        float64
	angle         float64
	matrix        ebiten.GeoM
	version       uint64
	matrixVersion uint64 // version when the matrix was last built
}

var Component = donburi.NewComponentType[Transform](Transform{
	sx: 1, sy: 1, version: 1,
})

func (t *Transform) Position() (x, y float64) {
	return t.x, t.y
}

func (t *Transform) Scale() (sx, sy float64) {
	return t.sx, t.sy
}

func (t *Transform) Origin() (ox, oy float64) {
	return t.ox, t.oy
}

func (t *Transform) Angle() float64 {
	return t.angle
}

// Matrix is only current after Update.
func (t *Transform) Matrix() ebiten.GeoM {
	return t.matrix
}

func (t *Transform) Version() uint64 {
	return t.version
}

func (t *Transform) SetPosition(x, y float64) {
	if t.x != x || t.y != y {
		t.x = x
		t.y = y
		t.version++
	}
}

func (t *Transform) SetScale(sx, sy float64) {
	if t.sx != sx || t.sy != sy {
		t.sx = sx
		t.sy = sy
		t.version++
	}
}

func (t *Transform) SetOrigin(ox, oy float64) {
	if t.ox != ox || t.oy != oy {
		t.ox = ox
		t.oy = oy
		t.version++
	}
}

func (t *Transform) SetAngle(angle float64) {
	if t.angle != angle {
		t.angle = angle
		t.version++
	}
}

// Update rebuilds matrices for transforms that changed.
func Update(w donburi.World) {
	for e := range Component.Iter(w) {
		t := Component.Get(e)
		if t.matrixVersion == t.version {
			continue
		}
		t.matrix = local(t)
		t.matrixVersion = t.version
	}
}

// local scales and rotates around the origin, then moves into position.
func local(t *Transform) ebiten.GeoM {
	var m ebiten.GeoM
	m.Translate(-t.ox, -t.oy)
	m.Scale(t.sx, t.sy)
	m.Rotate(t.angle)
	m.Translate(t.x+t.ox, t.y+t.oy)
	return m
}
