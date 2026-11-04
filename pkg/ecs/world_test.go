package ecs

import (
	"slices"
	"testing"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/yohamta/donburi"

	"github.com/adm87/finch/pkg/ecs/render"
	"github.com/adm87/finch/pkg/ecs/transform"
)

type recorder struct {
	drawn []*donburi.Entry
}

func (a *recorder) AppendJobs(e *donburi.Entry, _ *render.View, jobs []render.Job) []render.Job {
	return append(jobs, render.Job{Entry: e})
}

func (a *recorder) Draw(_ *ebiten.Image, _ *render.View, j *render.Job) {
	a.drawn = append(a.drawn, j.Entry)
}

func TestNewWorldPanicsOnNilPipeline(t *testing.T) {
	defer func() {
		if recover() == nil {
			t.Error("NewWorld(nil) did not panic")
		}
	}()
	NewWorld(nil)
}

func TestWorldSyncRebuildsTransforms(t *testing.T) {
	w := NewWorld(render.NewPipeline())
	d := w.Donburi()
	tr := transform.Component.Get(d.Entry(d.Create(transform.Component)))
	tr.SetPosition(3, 4)

	w.Sync()

	m := tr.Matrix()
	if x, y := m.Apply(0, 0); x != 3 || y != 4 {
		t.Errorf("matrix moved (0, 0) to (%v, %v), want (3, 4)", x, y)
	}
}

func TestWorldRenderDrawsOnlyRenderers(t *testing.T) {
	p := render.NewPipeline()
	a := &recorder{}
	kind := p.Register(a)

	w := NewWorld(p)
	d := w.Donburi()
	drawn := d.Entry(d.Create(render.Component))
	render.Component.Get(drawn).SetKind(kind)
	d.Create(transform.Component) // no Renderer, so never drawn

	w.Render(nil, render.View{})

	if want := []*donburi.Entry{drawn}; !slices.Equal(a.drawn, want) {
		t.Errorf("drawn = %v, want %v", a.drawn, want)
	}
}

func TestWorldRenderAllocs(t *testing.T) {
	p := render.NewPipeline()
	a := &recorder{}
	kind := p.Register(a)

	w := NewWorld(p)
	d := w.Donburi()
	for range 100 {
		render.Component.Get(d.Entry(d.Create(render.Component))).SetKind(kind)
	}

	frame := func() {
		a.drawn = a.drawn[:0]
		w.Render(nil, render.View{})
	}
	frame()

	if n := testing.AllocsPerRun(100, frame); n != 0 {
		t.Errorf("Render allocates %v times per frame", n)
	}
}
