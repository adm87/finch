package ecs

import (
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/yohamta/donburi"
	"github.com/yohamta/donburi/filter"

	"github.com/adm87/finch/pkg/ecs/rendering"
	"github.com/adm87/finch/pkg/ecs/transform"
)

// World owns the entities and draws them through a render pipeline.
type World struct {
	donburi   donburi.World
	pipeline  *rendering.Pipeline
	layered   *donburi.Query
	drawables []*donburi.Entry
}

func NewWorld(pipeline *rendering.Pipeline) *World {
	if pipeline == nil {
		panic("ecs: nil render pipeline")
	}
	return &World{
		pipeline: pipeline,
		donburi:  donburi.NewWorld(),
		layered:  donburi.NewQuery(filter.Contains(rendering.Component)),
	}
}

func (w *World) Donburi() donburi.World {
	return w.donburi
}

// Sync brings derived data, like transform matrices, up to date.
// Call it after the frame's updates.
func (w *World) Sync() {
	transform.Update(w.donburi)
}

// Render draws everything inside the view to target.
func (w *World) Render(target *ebiten.Image, view rendering.View) {
	w.drawables = w.drawables[:0]
	for e := range w.layered.Iter(w.donburi) {
		w.drawables = append(w.drawables, e)
	}
	w.pipeline.Draw(target, view, w.drawables)
}
