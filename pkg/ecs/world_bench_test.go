package ecs

import (
	"fmt"
	"testing"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/yohamta/donburi"

	"github.com/adm87/finch/pkg/ecs/render"
	"github.com/adm87/finch/pkg/ecs/transform"
)

type noop struct{}

func (noop) AppendJobs(e *donburi.Entry, _ *render.View, jobs []render.Job) []render.Job {
	return append(jobs, render.Job{Entry: e})
}

func (noop) Draw(*ebiten.Image, *render.View, *render.Job) {}

func BenchmarkWorldRender(b *testing.B) {
	for _, n := range []int{100, 1_000, 10_000} {
		b.Run(fmt.Sprintf("entities=%d", n), func(b *testing.B) {
			p := render.NewPipeline()
			kind := p.Register(noop{})

			w := NewWorld(p)
			d := w.Donburi()
			for range n {
				render.Component.Get(d.Entry(d.Create(render.Component))).SetKind(kind)
			}

			b.ReportAllocs()
			for b.Loop() {
				w.Render(nil, render.View{})
			}
		})
	}
}

func BenchmarkWorldSync(b *testing.B) {
	for _, n := range []int{100, 1_000, 10_000} {
		b.Run(fmt.Sprintf("entities=%d", n), func(b *testing.B) {
			w := NewWorld(render.NewPipeline())
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
