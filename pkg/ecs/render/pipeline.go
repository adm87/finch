// Package render turns entities into draw commands and draws them
// in layer and depth order.
//
// Pipeline runs its adapters over the visible entries, then draws
// the queued commands in order.
package render

import (
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/yohamta/donburi"
)

// Adapter knows how to draw entities of a certain type, and enqueues them in the render queue.
type Adapter interface {
	// Enqueue adds draw commands for the given entries to the queue.
	Enqueue(queue *Queue, entries []*donburi.Entry, view View)
}

// AdapterFunc adapts a function to an Adapter.
type AdapterFunc func(queue *Queue, entries []*donburi.Entry, view View)

func (f AdapterFunc) Enqueue(queue *Queue, entries []*donburi.Entry, view View) {
	f(queue, entries, view)
}

// Pipeline draws entities in order of layer and depth, using the adapter that knows how to draw each one.
type Pipeline struct {
	adapters []Adapter
	visible  []*donburi.Entry
	queue    Queue
}

func NewPipeline(adapters ...Adapter) *Pipeline {
	for _, a := range adapters {
		if a == nil {
			panic("render: nil adapter")
		}
	}
	return &Pipeline{adapters: adapters}
}

func (p *Pipeline) Draw(target *ebiten.Image, view View, entries []*donburi.Entry) {
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
