package sprite_test

import (
	"fmt"
	"testing"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/yohamta/donburi"

	"github.com/adm87/finch/pkg/ecs/render"
	"github.com/adm87/finch/pkg/ecs/sprite"
	"github.com/adm87/finch/pkg/ecs/transform"
	"github.com/adm87/finch/pkg/geom"
)

// Every other sprite is placed off-screen, so half the entities are culled.
func BenchmarkAdapterAppendJobs(b *testing.B) {
	for _, n := range []int{100, 1_000, 10_000} {
		b.Run(fmt.Sprintf("entities=%d", n), func(b *testing.B) {
			w := donburi.NewWorld()
			a := sprite.NewAdapter(images{ebiten.NewImage(10, 10)})
			view := &render.View{Bounds: geom.Rect{Width: 400, Height: 300}}

			entries := make([]*donburi.Entry, n)
			for i := range entries {
				x := float64(i % 400)
				if i%2 == 1 {
					x += 1000
				}
				e := newSprite(w, 0, x, float64(i%300))
				transform.Component.Get(e).SetAngle(float64(i))
				entries[i] = e
			}
			transform.Update(w)

			jobs := make([]render.Job, 0, n)
			b.ReportAllocs()
			for b.Loop() {
				jobs = jobs[:0]
				for _, e := range entries {
					jobs = a.AppendJobs(e, view, jobs)
				}
			}
		})
	}
}
