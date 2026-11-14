package rendering

import (
	"fmt"
	"testing"

	"github.com/yohamta/donburi"
)

// noop measures the pipeline on its own: no commands, nothing drawn.
type noop struct{}

func (noop) Enqueue(*Queue, []*donburi.Entry, View) {}

func BenchmarkPipelineDraw(b *testing.B) {
	for _, n := range []int{100, 1_000, 10_000} {
		b.Run(fmt.Sprintf("entities=%d", n), func(b *testing.B) {
			w := donburi.NewWorld()
			p := NewPipeline(noop{})

			entries := make([]*donburi.Entry, n)
			for i := range entries {
				entries[i] = newEntry(w, 0, 0)
			}

			b.ReportAllocs()
			for b.Loop() {
				p.Draw(nil, View{}, entries)
			}
		})
	}
}
