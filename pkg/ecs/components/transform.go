package components

import (
	"github.com/yohamta/donburi"
)

type Transform struct {
	x, y     float64
	ox, oy   float64
	sx, sy   float64
	rotation float64
	isDirty  bool
}

var TransformComponent = donburi.NewComponentType[Transform](Transform{
	sx:      1,
	sy:      1,
	isDirty: true,
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
	t.isDirty = true
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
	t.isDirty = true
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
	t.isDirty = true
}

func (t *Transform) Rotation() float64 {
	return t.rotation
}

func (t *Transform) SetRotation(rotation float64) {
	if t.rotation == rotation {
		return
	}
	t.rotation = rotation
	t.isDirty = true
}
