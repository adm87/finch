package render

import "github.com/yohamta/donburi"

// Kind picks which adapter draws an entity. Zero means none.
type Kind int

type Renderer struct {
	kind    Kind
	visible bool
	layer   int
	depth   int
}

var Component = donburi.NewComponentType[Renderer](Renderer{
	visible: true,
})

func (r *Renderer) Kind() Kind {
	return r.kind
}

func (r *Renderer) Visible() bool {
	return r.visible
}

func (r *Renderer) Layer() int {
	return r.layer
}

func (r *Renderer) Depth() int {
	return r.depth
}

func (r *Renderer) SetKind(kind Kind) {
	r.kind = kind
}

func (r *Renderer) SetVisible(visible bool) {
	r.visible = visible
}

// SetLayer sets the draw layer. Layers are clamped to int8 when sorting.
func (r *Renderer) SetLayer(layer int) {
	r.layer = layer
}

// SetDepth orders entities within a layer. Depths are clamped to int16 when sorting.
func (r *Renderer) SetDepth(depth int) {
	r.depth = depth
}
