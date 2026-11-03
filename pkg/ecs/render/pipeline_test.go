package render

import (
	"slices"
	"testing"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/yohamta/donburi"
)

type recorder struct {
	drawn []*donburi.Entry
}

func (a *recorder) AppendJobs(e *donburi.Entry, _ *View, jobs []Job) []Job {
	r := Component.Get(e)
	return append(jobs, Job{Entry: e, Layer: r.Layer(), Depth: r.Depth()})
}

func (a *recorder) Draw(_ *ebiten.Image, _ *View, j *Job) {
	a.drawn = append(a.drawn, j.Entry)
}

func newEntry(w donburi.World, kind Kind, layer, depth int) *donburi.Entry {
	e := w.Entry(w.Create(Component))
	r := Component.Get(e)
	r.SetKind(kind)
	r.SetLayer(layer)
	r.SetDepth(depth)
	return e
}

func TestPipelineDrawOrder(t *testing.T) {
	w := donburi.NewWorld()
	p := NewPipeline()
	a := &recorder{}
	kind := p.Register(a)

	top := newEntry(w, kind, 1, 0)
	front := newEntry(w, kind, 0, 5)
	back := newEntry(w, kind, 0, -5)
	under := newEntry(w, kind, -1, 100)

	p.Draw(nil, View{}, []*donburi.Entry{top, front, back, under})

	want := []*donburi.Entry{under, back, front, top}
	if !slices.Equal(a.drawn, want) {
		t.Errorf("draw order = %v, want %v", a.drawn, want)
	}
}

func TestPipelineDrawSkips(t *testing.T) {
	w := donburi.NewWorld()
	p := NewPipeline()
	a := &recorder{}
	kind := p.Register(a)

	shown := newEntry(w, kind, 0, 0)
	hidden := newEntry(w, kind, 0, 0)
	Component.Get(hidden).SetVisible(false)
	none := newEntry(w, 0, 0, 0)

	p.Draw(nil, View{}, []*donburi.Entry{shown, hidden, none})

	if want := []*donburi.Entry{shown}; !slices.Equal(a.drawn, want) {
		t.Errorf("drawn = %v, want %v", a.drawn, want)
	}
}

func TestPipelineDrawAllocs(t *testing.T) {
	w := donburi.NewWorld()
	p := NewPipeline()
	a := &recorder{}
	kind := p.Register(a)

	entries := make([]*donburi.Entry, 1000)
	for i := range entries {
		entries[i] = newEntry(w, kind, i%3, -i)
	}

	frame := func() {
		a.drawn = a.drawn[:0]
		p.Draw(nil, View{}, entries)
	}
	frame()

	if n := testing.AllocsPerRun(100, frame); n != 0 {
		t.Errorf("Draw allocates %v times per frame", n)
	}
}

func TestKeyOrder(t *testing.T) {
	tests := []struct {
		name string
		a, b Job
	}{
		{"layer before depth", Job{Layer: -1, Depth: 100}, Job{Layer: 0, Depth: -100}},
		{"negative depth", Job{Depth: -1}, Job{Depth: 0}},
		{"kind within depth", Job{kind: 1}, Job{kind: 2}},
		{"clamped layer", Job{Layer: -1000}, Job{Layer: 1000}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if key(&tt.a, 1) >= key(&tt.b, 0) {
				t.Errorf("key(%+v) >= key(%+v)", tt.a, tt.b)
			}
		})
	}
}
