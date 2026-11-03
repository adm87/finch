package transform

import (
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/yohamta/donburi"
)

type Transform struct {
	x, y          float64
	sx, sy        float64
	ox, oy        float64
	angle         float64
	matrix        ebiten.GeoM
	version       uint64
	matrixVersion uint64 // version the matrix was built from
}

var Component = donburi.NewComponentType[Transform](Transform{
	sx: 1, sy: 1, version: 1,
})

func Get(e *donburi.Entry) (*Transform, bool) {
	if e.HasComponent(Component) {
		return Component.Get(e), true
	}
	return nil, false
}

// Update rebuilds the matrix of every transform that changed since it was last built.
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

// Matrix is the transform as of the last Update.
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

// local scales and rotates around the origin, then moves to the position.
func local(t *Transform) ebiten.GeoM {
	var m ebiten.GeoM
	m.Translate(-t.ox, -t.oy)
	m.Scale(t.sx, t.sy)
	m.Rotate(t.angle)
	m.Translate(t.x+t.ox, t.y+t.oy)
	return m
}
