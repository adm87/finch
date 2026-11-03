package ecs

import (
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/yohamta/donburi"
	"github.com/yohamta/donburi/filter"

	"github.com/adm87/finch/pkg/ecs/render"
	"github.com/adm87/finch/pkg/ecs/transform"
)

type World struct {
	donburi   donburi.World
	pipeline  *render.Pipeline
	renderers *donburi.Query
	drawables []*donburi.Entry
}

func NewWorld(pipeline *render.Pipeline) *World {
	if pipeline == nil {
		panic("ecs: nil render pipeline")
	}
	return &World{
		pipeline:  pipeline,
		donburi:   donburi.NewWorld(),
		renderers: donburi.NewQuery(filter.Contains(render.Component)),
	}
}

func (w *World) Donburi() donburi.World {
	return w.donburi
}

// Sync rebuilds derived data, such as transform matrices, after the frame's updates.
func (w *World) Sync() {
	transform.Update(w.donburi)
}

// Render draws the entities inside the view to target.
func (w *World) Render(target *ebiten.Image, view render.View) {
	w.drawables = w.drawables[:0]
	for e := range w.renderers.Iter(w.donburi) {
		w.drawables = append(w.drawables, e)
	}
	w.pipeline.Draw(target, view, w.drawables)
}
