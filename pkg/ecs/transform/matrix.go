package transform

import (
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/yohamta/donburi"
)

type Matrix struct {
	geoM ebiten.GeoM
}

var MatrixComponent = donburi.NewComponentType[Matrix]()

func GetMatrix(entry *donburi.Entry) (*Matrix, bool) {
	if !entry.HasComponent(MatrixComponent) {
		return nil, false
	}

	return MatrixComponent.Get(entry), true
}

func GetGeoM(entry *donburi.Entry) ebiten.GeoM {
	m, mOk := GetMatrix(entry)
	if !mOk {
		return ebiten.GeoM{}
	}

	t, tOk := GetTransform(entry)
	if !tOk {
		return m.geoM
	}

	return GetGeoMOf(t, m)
}

func GetGeoMOf(t *Transform, m *Matrix) ebiten.GeoM {
	if t.isDirty {
		m.geoM.Reset()
		m.geoM.Translate(-t.ox, -t.oy)
		m.geoM.Scale(t.sx, t.sy)
		m.geoM.Rotate(t.rotation)
		m.geoM.Translate(t.x, t.y)
		t.isDirty = false
	}

	return m.geoM
}
