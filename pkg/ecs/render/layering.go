package render

import "github.com/yohamta/donburi"

// Layering decides whether an entity is drawn and where it sits in the draw order.
type Layering struct {
	visible bool
	layer   int16
	depth   int16
}

var Component = donburi.NewComponentType[Layering](Layering{
	visible: true,
})

func (l *Layering) Visible() bool {
	return l.visible
}

func (l *Layering) Layer() int16 {
	return l.layer
}

func (l *Layering) Depth() int16 {
	return l.depth
}

func (l *Layering) SetVisible(visible bool) {
	l.visible = visible
}

func (l *Layering) SetLayer(layer int16) {
	l.layer = layer
}

func (l *Layering) SetDepth(depth int16) {
	l.depth = depth
}
