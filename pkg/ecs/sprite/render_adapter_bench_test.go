package sprite_test

import (
	"fmt"
	"testing"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/yohamta/donburi"

	"github.com/adm87/finch/pkg/ecs/render"
	"github.com/adm87/finch/pkg/ecs/sprite"
	"github.com/adm87/finch/pkg/ecs/transform"
)

func BenchmarkRenderAdapterEnqueue(b *testing.B) {
	for _, n := range []int{100, 1_000, 10_000} {
		b.Run(fmt.Sprintf("entities=%d", n), func(b *testing.B) {
			w := donburi.NewWorld()
			a := sprite.NewRenderAdapter(images{ebiten.NewImage(10, 10)})

			entries := make([]*donburi.Entry, n)
			for i := range entries {
				e := w.Entry(w.Create(render.Component, sprite.Component, transform.Component))
				sprite.Component.Get(e).SetHandle(1)
				t := transform.Component.Get(e)
				t.SetPosition(float64(i%400), float64(i%300))
				t.SetAngle(float64(i))
				entries[i] = e
			}
			transform.Update(w)

			var q render.Queue
			b.ReportAllocs()
			for b.Loop() {
				a.Enqueue(&q, entries, render.View{})
				q.Reset()
			}
		})
	}
}
