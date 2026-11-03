package render

import "github.com/yohamta/donburi"

// Kind routes an entity to a registered Adapter. The zero Kind has no adapter.
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

func Get(e *donburi.Entry) (*Renderer, bool) {
	if e.HasComponent(Component) {
		return Component.Get(e), true
	}
	return nil, false
}

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

// SetLayer sets the draw layer. Layers outside the int8 range are clamped when sorted.
func (r *Renderer) SetLayer(layer int) {
	r.layer = layer
}

// SetDepth sets the order within a layer. Depths outside the int16 range are clamped when sorted.
func (r *Renderer) SetDepth(depth int) {
	r.depth = depth
}
