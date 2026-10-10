package bounds

import (
	"github.com/adm87/finch/pkg/geom"
	"github.com/yohamta/donburi"
)

type Bounds struct {
	rect geom.Rect
}

var ComponentType = donburi.NewComponentType[Bounds]()

func (b *Bounds) Rect() geom.Rect {
	return b.rect
}

func (b *Bounds) SetRect(rect geom.Rect) {
	b.rect = rect
}
