package ecs

import (
	"slices"
	"testing"

	"github.com/yohamta/donburi"

	"github.com/adm87/finch/pkg/ecs/render"
	"github.com/adm87/finch/pkg/ecs/transform"
)

type recorder struct {
	got []*donburi.Entry
}

func (a *recorder) Enqueue(_ *render.Queue, entries []*donburi.Entry, _ render.View) {
	a.got = append(a.got, entries...)
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

func TestWorldRenderDrawsOnlyLayered(t *testing.T) {
	a := &recorder{}
	w := NewWorld(render.NewPipeline(a))
	d := w.Donburi()
	drawn := d.Entry(d.Create(render.Component))
	d.Create(transform.Component) // no Layering, so never drawn

	w.Render(nil, render.View{})

	if want := []*donburi.Entry{drawn}; !slices.Equal(a.got, want) {
		t.Errorf("adapter got %v, want %v", a.got, want)
	}
}

func TestWorldRenderAllocs(t *testing.T) {
	a := &recorder{}
	w := NewWorld(render.NewPipeline(a))
	d := w.Donburi()
	for range 100 {
		d.Create(render.Component)
	}

	frame := func() {
		a.got = a.got[:0]
		w.Render(nil, render.View{})
	}
	frame()

	if n := testing.AllocsPerRun(100, frame); n != 0 {
		t.Errorf("Render allocates %v times per frame", n)
	}
}
