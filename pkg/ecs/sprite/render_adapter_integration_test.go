//go:build integration

package sprite_test

import (
	"image/color"
	"testing"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/yohamta/donburi"

	"github.com/adm87/finch/pkg/ecs/render"
	"github.com/adm87/finch/pkg/ecs/sprite"
	"github.com/adm87/finch/pkg/ecs/transform"
	"github.com/adm87/finch/pkg/geom"
	"github.com/adm87/finch/pkg/internal/ebitentest"
)

const size = 16

var (
	blank = color.RGBA{}
	white = color.RGBA{R: 255, G: 255, B: 255, A: 255}
	red   = color.RGBA{R: 255, A: 255}
)

func TestIntegrationRenderAdapter(t *testing.T) {
	tests := []struct {
		name  string
		setup func(e *donburi.Entry, view *render.View)
		x, y  int
		want  color.RGBA
	}{
		{"draws at position", nil, 4, 4, white},
		{"nothing before position", nil, 3, 3, blank},
		{"anchor", func(e *donburi.Entry, _ *render.View) {
			sprite.Component.Get(e).SetAnchor(0.5, 0.5)
		}, 3, 3, white},
		{"color", func(e *donburi.Entry, _ *render.View) {
			var c ebiten.ColorScale
			c.Scale(1, 0, 0, 1)
			sprite.Component.Get(e).SetColor(c)
		}, 4, 4, red},
		{"view matrix", func(_ *donburi.Entry, view *render.View) {
			view.Matrix.Translate(2, 0)
		}, 6, 4, white},
		{"culled outside view", func(_ *donburi.Entry, view *render.View) {
			view.Bounds = geom.Rect{X: 8, Y: 8, Width: 8, Height: 8}
		}, 4, 4, blank},
		{"image not loaded", func(e *donburi.Entry, _ *render.View) {
			sprite.Component.Get(e).SetHandle(0)
		}, 4, 4, blank},
		{"no transform", func(e *donburi.Entry, _ *render.View) {
			e.RemoveComponent(transform.Component)
		}, 0, 0, blank},
	}

	got := make([]color.Color, len(tests))
	ebitentest.InFrame(t, func() {
		square := ebiten.NewImage(2, 2)
		square.Fill(white)
		p := render.NewPipeline(sprite.NewRenderAdapter(images{square}))

		for i, tt := range tests {
			w := donburi.NewWorld()
			e := w.Entry(w.Create(render.Component, sprite.Component, transform.Component))
			sprite.Component.Get(e).SetHandle(1)
			transform.Component.Get(e).SetPosition(4, 4)
			view := render.View{Bounds: geom.Rect{Width: size, Height: size}}
			if tt.setup != nil {
				tt.setup(e, &view)
			}
			transform.Update(w)

			target := ebiten.NewImage(size, size)
			p.Draw(target, view, []*donburi.Entry{e})
			got[i] = target.At(tt.x, tt.y)
		}
	})

	for i, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got[i] != tt.want {
				t.Errorf("pixel (%d, %d) = %v, want %v", tt.x, tt.y, got[i], tt.want)
			}
		})
	}
}
