package render

import (
	"fmt"
	"testing"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/yohamta/donburi"
)

// noop measures the pipeline on its own: one job per entry, nothing drawn.
type noop struct{}

func (noop) AppendJobs(e *donburi.Entry, _ *View, jobs []Job) []Job {
	r := Component.Get(e)
	return append(jobs, Job{Entry: e, Layer: r.Layer(), Depth: r.Depth()})
}

func (noop) Draw(*ebiten.Image, *View, *Job) {}

func benchEntries(n int, order func(i int) (layer, depth int)) (*Pipeline, []*donburi.Entry) {
	w := donburi.NewWorld()
	p := NewPipeline()
	kind := p.Register(noop{})

	entries := make([]*donburi.Entry, n)
	for i := range entries {
		layer, depth := order(i)
		entries[i] = newEntry(w, kind, layer, depth)
	}
	return p, entries
}

func BenchmarkPipelineDraw(b *testing.B) {
	orders := []struct {
		name  string
		order func(i int) (layer, depth int)
	}{
		{"uniform", func(int) (int, int) { return 0, 0 }},
		{"sorted", func(i int) (int, int) { return 0, i }},
		{"mixed", func(i int) (int, int) { return i % 4, (i * 7919) % 1000 }},
	}
	for _, n := range []int{100, 1_000, 10_000} {
		for _, o := range orders {
			b.Run(fmt.Sprintf("entities=%d/%s", n, o.name), func(b *testing.B) {
				p, entries := benchEntries(n, o.order)
				b.ReportAllocs()
				for b.Loop() {
					p.Draw(nil, View{}, entries)
				}
			})
		}
	}
}
