package renderer

import "github.com/yohamta/donburi"

type Renderer struct {
	id      uint64
	layer   int16
	depth   int16
	visible bool
}

var ComponentType = donburi.NewComponentType[Renderer](Renderer{
	visible: true,
})

func (r *Renderer) ID() uint64 {
	return r.id
}

func (r *Renderer) SetID(id uint64) {
	r.id = id
}

func (r *Renderer) Layer() int16 {
	return r.layer
}

func (r *Renderer) SetLayer(layer int16) {
	r.layer = layer
}

func (r *Renderer) Depth() int16 {
	return r.depth
}

func (r *Renderer) SetDepth(depth int16) {
	r.depth = depth
}

func (r *Renderer) Visible() bool {
	return r.visible
}

func (r *Renderer) SetVisible(visible bool) {
	r.visible = visible
}
