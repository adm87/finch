package rendering

import (
	"math"
	"slices"
	"testing"

	"github.com/yohamta/donburi"

	"github.com/adm87/finch/pkg/ecs/bounds"
	"github.com/adm87/finch/pkg/ecs/transform"
	"github.com/adm87/finch/pkg/geom"
)

var other = donburi.NewTag()

type recorder struct {
	got []*donburi.Entry
}

func (a *recorder) Enqueue(_ *Queue, entries []*donburi.Entry, _ View) {
	a.got = append(a.got, entries...)
}

func TestPipelineDrawSkipsHidden(t *testing.T) {
	w := donburi.NewWorld()
	a := &recorder{}
	p := NewPipeline(a)

	shown := newEntry(w, 0, 0)
	hidden := newEntry(w, 0, 0)
	Component.Get(hidden).SetVisible(false)
	none := w.Entry(w.Create(other)) // no Layering

	p.Draw(nil, View{}, []*donburi.Entry{shown, hidden, none})

	if want := []*donburi.Entry{shown}; !slices.Equal(a.got, want) {
		t.Errorf("adapter got %v, want %v", a.got, want)
	}
}

func TestPipelineDrawCullsOffscreen(t *testing.T) {
	w := donburi.NewWorld()
	a := &recorder{}
	p := NewPipeline(a)

	onscreen := newBoundedEntry(w, 50, 50)
	edge := newBoundedEntry(w, 105, 50) // half inside the view
	offscreen := newBoundedEntry(w, 500, 50)
	unbounded := newEntry(w, 0, 0) // never culled
	transform.Update(w)

	view := View{Bounds: geom.Rect{Width: 100, Height: 100}}
	p.Draw(nil, view, []*donburi.Entry{onscreen, edge, offscreen, unbounded})

	if want := []*donburi.Entry{onscreen, edge, unbounded}; !slices.Equal(a.got, want) {
		t.Errorf("adapter got %v, want %v", a.got, want)
	}
}

func TestPipelineDrawRunsEveryAdapter(t *testing.T) {
	w := donburi.NewWorld()
	var order []string
	first := AdapterFunc(func(*Queue, []*donburi.Entry, View) { order = append(order, "first") })
	second := AdapterFunc(func(*Queue, []*donburi.Entry, View) { order = append(order, "second") })
	p := NewPipeline(first, second)

	p.Draw(nil, View{}, []*donburi.Entry{newEntry(w, 0, 0)})

	if want := []string{"first", "second"}; !slices.Equal(order, want) {
		t.Errorf("adapters ran %v, want %v", order, want)
	}
}

func TestPipelineDrawAllocs(t *testing.T) {
	w := donburi.NewWorld()
	a := &recorder{}
	p := NewPipeline(a)

	entries := make([]*donburi.Entry, 1000)
	for i := range entries {
		entries[i] = newEntry(w, int16(i%3), int16(-i))
	}

	frame := func() {
		a.got = a.got[:0]
		p.Draw(nil, View{}, entries)
	}
	frame()

	if n := testing.AllocsPerRun(100, frame); n != 0 {
		t.Errorf("Draw allocates %v times per frame", n)
	}
}

func TestNewPipelinePanicsOnNilAdapter(t *testing.T) {
	defer func() {
		if recover() == nil {
			t.Error("NewPipeline(nil) did not panic")
		}
	}()
	NewPipeline(nil)
}

func TestKeyOrder(t *testing.T) {
	type pos struct {
		layer, depth int16
		index        int
	}
	tests := []struct {
		name string
		a, b pos
	}{
		{"layer before depth", pos{-1, 100, 0}, pos{0, -100, 0}},
		{"negative layer", pos{-1, 0, 0}, pos{1, 0, 0}},
		{"negative depth", pos{0, -1, 0}, pos{0, 1, 0}},
		{"index breaks ties", pos{0, 0, 0}, pos{0, 0, 1}},
		{"extremes", pos{math.MinInt16, math.MaxInt16, 1}, pos{math.MaxInt16, math.MinInt16, 0}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if key(tt.a.layer, tt.a.depth, tt.a.index) >= key(tt.b.layer, tt.b.depth, tt.b.index) {
				t.Errorf("key(%+v) >= key(%+v)", tt.a, tt.b)
			}
		})
	}
}

func newEntry(w donburi.World, layer, depth int16) *donburi.Entry {
	e := w.Entry(w.Create(Component))
	r := Component.Get(e)
	r.SetLayer(layer)
	r.SetDepth(depth)
	return e
}

func newBoundedEntry(w donburi.World, x, y float64) *donburi.Entry {
	e := w.Entry(w.Create(Component, transform.Component, bounds.Component))
	transform.Component.Get(e).SetPosition(x, y)
	bounds.Component.Get(e).SetLocal(geom.Rect{X: -8, Y: -8, Width: 16, Height: 16})
	return e
}
