package render

import "github.com/yohamta/donburi"

type Renderer struct {
	visible bool
	layer   int16
	depth   int16
}

var Component = donburi.NewComponentType[Renderer](Renderer{
	visible: true,
})

func (r *Renderer) Visible() bool {
	return r.visible
}

func (r *Renderer) Layer() int16 {
	return r.layer
}

func (r *Renderer) Depth() int16 {
	return r.depth
}

func (r *Renderer) SetVisible(visible bool) {
	r.visible = visible
}

func (r *Renderer) SetLayer(layer int16) {
	r.layer = layer
}

func (r *Renderer) SetDepth(depth int16) {
	r.depth = depth
}
