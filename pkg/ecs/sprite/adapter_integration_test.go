//go:build integration

package sprite_test

import (
	"testing"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/yohamta/donburi"

	"github.com/adm87/finch/pkg/ecs/render"
	"github.com/adm87/finch/pkg/ecs/sprite"
	"github.com/adm87/finch/pkg/ecs/transform"
	"github.com/adm87/finch/pkg/engine"
	"github.com/adm87/finch/pkg/geom"
)

// Ebitengine copies draws made outside a running game, so allocations are only measured inside a frame.
func TestIntegrationAdapterDrawAllocs(t *testing.T) {
	w := donburi.NewWorld()
	p := render.NewPipeline()
	kind := p.Register(sprite.NewAdapter(images{ebiten.NewImage(4, 4)}))

	entries := make([]*donburi.Entry, 1000)
	for i := range entries {
		e := w.Entry(w.Create(render.Component, sprite.Component, transform.Component))
		render.Component.Get(e).SetKind(kind)
		transform.Component.Get(e).SetPosition(float64(i%64), float64(i/64))
		entries[i] = e
	}
	transform.Update(w)

	view := render.View{Bounds: geom.Rect{Width: 64, Height: 64}}

	var allocs float64
	var measured bool
	err := engine.Run(
		engine.WithUpdater(engine.UpdaterFunc(func(float64) error {
			if measured {
				return ebiten.Termination
			}
			return nil
		})),
		engine.WithDrawer(engine.DrawerFunc(func(target *ebiten.Image) {
			if measured {
				return
			}
			for range 10 {
				p.Draw(target, view, entries)
			}
			allocs = testing.AllocsPerRun(100, func() { p.Draw(target, view, entries) })
			measured = true
		})),
	)
	if err != nil {
		t.Fatalf("Run() = %v", err)
	}
	if allocs != 0 {
		t.Errorf("Draw allocates %v times per frame", allocs)
	}
}
