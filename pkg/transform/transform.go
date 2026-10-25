package transform

import (
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/yohamta/donburi"
)

type Transform struct {
	x, y     float64
	ox, oy   float64
	sx, sy   float64
	rotation float64
	geoM     ebiten.GeoM
	// geoMVersion is the version geoM was built for.
	geoMVersion uint64
	version     uint64
}

var TransformComponent = donburi.NewComponentType[Transform](Transform{
	sx: 1,
	sy: 1,
	// Starting at 1 leaves the cached geoM, built for version 0, out of date.
	version: 1,
})

func GetTransform(entry *donburi.Entry) (*Transform, bool) {
	if !entry.HasComponent(TransformComponent) {
		return nil, false
	}

	return TransformComponent.Get(entry), true
}

func (t *Transform) Position() (float64, float64) {
	return t.x, t.y
}

func (t *Transform) SetPosition(x, y float64) {
	if t.x == x && t.y == y {
		return
	}
	t.x = x
	t.y = y
	t.version++
}

func (t *Transform) Origin() (float64, float64) {
	return t.ox, t.oy
}

func (t *Transform) SetOrigin(ox, oy float64) {
	if t.ox == ox && t.oy == oy {
		return
	}
	t.ox = ox
	t.oy = oy
	t.version++
}

func (t *Transform) Scale() (float64, float64) {
	return t.sx, t.sy
}

func (t *Transform) SetScale(sx, sy float64) {
	if t.sx == sx && t.sy == sy {
		return
	}
	t.sx = sx
	t.sy = sy
	t.version++
}

func (t *Transform) Rotation() float64 {
	return t.rotation
}

func (t *Transform) SetRotation(rotation float64) {
	if t.rotation == rotation {
		return
	}
	t.rotation = rotation
	t.version++
}

// Version changes whenever the position, origin, scale or rotation changes.
func (t *Transform) Version() uint64 {
	return t.version
}

func (t *Transform) GeoM() ebiten.GeoM {
	if t.geoMVersion != t.version {
		t.geoM.Reset()
		t.geoM.Translate(-t.ox, -t.oy)
		t.geoM.Scale(t.sx, t.sy)
		t.geoM.Rotate(t.rotation)
		t.geoM.Translate(t.x, t.y)
		t.geoMVersion = t.version
	}
	return t.geoM
}
