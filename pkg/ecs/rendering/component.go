package rendering

import "github.com/yohamta/donburi"

// Rendering decides whether an entity is drawn and where it sits in the draw order.
type Rendering struct {
	visible bool
	layer   int16
	depth   int16
}

var Component = donburi.NewComponentType[Rendering](Rendering{
	visible: true,
})

func (r *Rendering) Visible() bool {
	return r.visible
}

func (r *Rendering) Layer() int16 {
	return r.layer
}

func (r *Rendering) Depth() int16 {
	return r.depth
}

func (r *Rendering) SetVisible(visible bool) {
	r.visible = visible
}

func (r *Rendering) SetLayer(layer int16) {
	r.layer = layer
}

func (r *Rendering) SetDepth(depth int16) {
	r.depth = depth
}
