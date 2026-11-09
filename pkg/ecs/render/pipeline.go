package render

import (
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/yohamta/donburi"
)

// Adapter turns the entities it knows how to draw into draw commands.
type Adapter interface {
	// Enqueue pushes commands for any entries it can draw.
	Enqueue(queue *Queue, entries []*donburi.Entry, view View)
}

// AdapterFunc adapts a function to an Adapter.
type AdapterFunc func(queue *Queue, entries []*donburi.Entry, view View)

func (f AdapterFunc) Enqueue(queue *Queue, entries []*donburi.Entry, view View) {
	f(queue, entries, view)
}

// Pipeline draws entities through its adapters, in layer and depth order.
type Pipeline struct {
	adapters []Adapter
	visible  []*donburi.Entry
	queue    Queue
}

// NewPipeline runs adapters in the order given. It panics on a nil adapter.
func NewPipeline(adapters ...Adapter) *Pipeline {
	for _, a := range adapters {
		if a == nil {
			panic("render: nil adapter")
		}
	}
	return &Pipeline{adapters: adapters}
}

// Draw hands each adapter the entries with a visible Layering, then draws
// what they queued. The queue is cleared afterwards, even if an adapter panics.
func (p *Pipeline) Draw(target *ebiten.Image, view View, entries []*donburi.Entry) {
	defer p.queue.Reset()

	p.visible = p.visible[:0]

	for _, e := range entries {
		if !e.HasComponent(Component) || !Component.Get(e).Visible() {
			continue
		}
		p.visible = append(p.visible, e)
	}

	for _, a := range p.adapters {
		a.Enqueue(&p.queue, p.visible, view)
	}

	p.queue.flush(target)
}
