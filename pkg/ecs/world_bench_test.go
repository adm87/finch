package ecs

import (
	"fmt"
	"testing"

	"github.com/yohamta/donburi"

	"github.com/adm87/finch/pkg/ecs/rendering"
	"github.com/adm87/finch/pkg/ecs/transform"
)

type noop struct{}

func (noop) Enqueue(*rendering.Queue, []*donburi.Entry, rendering.View) {}

func BenchmarkWorldRender(b *testing.B) {
	for _, n := range []int{100, 1_000, 10_000} {
		b.Run(fmt.Sprintf("entities=%d", n), func(b *testing.B) {
			w := NewWorld(rendering.NewPipeline(noop{}))
			d := w.Donburi()
			for range n {
				d.Create(rendering.Component)
			}

			b.ReportAllocs()
			for b.Loop() {
				w.Render(nil, rendering.View{})
			}
		})
	}
}

func BenchmarkWorldSync(b *testing.B) {
	for _, n := range []int{100, 1_000, 10_000} {
		b.Run(fmt.Sprintf("entities=%d", n), func(b *testing.B) {
			w := NewWorld(rendering.NewPipeline())
			d := w.Donburi()
			transforms := make([]*transform.Transform, n)
			for i := range transforms {
				transforms[i] = transform.Component.Get(d.Entry(d.Create(transform.Component)))
			}

			b.ReportAllocs()
			var x float64
			for b.Loop() {
				x++
				for _, t := range transforms {
					t.SetPosition(x, 0)
				}
				w.Sync()
			}
		})
	}
}
