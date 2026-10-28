package rendering

import "github.com/yohamta/donburi"

type Renderer struct {
	rendererID AdapterID
	depth      int32
	layer      int16
	visible    bool
}

var RendererComponent = donburi.NewComponentType[Renderer](Renderer{
	visible: true,
})

func GetRenderer(entry *donburi.Entry) (*Renderer, bool) {
	if !entry.HasComponent(RendererComponent) {
		return nil, false
	}
	return RendererComponent.Get(entry), true
}

func (r *Renderer) RendererID() AdapterID {
	return r.rendererID
}

func (r *Renderer) SetRendererID(id AdapterID) {
	r.rendererID = id
}

func (r *Renderer) Layer() int16 {
	return r.layer
}

func (r *Renderer) SetLayer(layer int16) {
	r.layer = layer
}

func (r *Renderer) Depth() int32 {
	return r.depth
}

func (r *Renderer) SetDepth(depth int32) {
	r.depth = depth
}

func (r *Renderer) Visible() bool {
	return r.visible
}

func (r *Renderer) SetVisible(visible bool) {
	r.visible = visible
}
