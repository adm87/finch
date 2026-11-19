//go:build integration

package sprite_test

import (
	"image"
	"image/color"
	"testing"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/yohamta/donburi"

	"github.com/adm87/finch/pkg/ecs/rendering"
	"github.com/adm87/finch/pkg/ecs/sprite"
	"github.com/adm87/finch/pkg/ecs/transform"
	"github.com/adm87/finch/pkg/images"
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
		setup func(e *donburi.Entry, view *rendering.View)
		x, y  int
		want  color.RGBA
	}{
		{"draws at position", nil, 4, 4, white},
		{"nothing before position", nil, 3, 3, blank},
		{"anchor", func(e *donburi.Entry, _ *rendering.View) {
			sprite.Component.Get(e).SetAnchor(0.5, 0.5)
		}, 3, 3, white},
		{"color", func(e *donburi.Entry, _ *rendering.View) {
			var c ebiten.ColorScale
			c.Scale(1, 0, 0, 1)
			sprite.Component.Get(e).SetColor(c)
		}, 4, 4, red},
		{"frame offset", func(e *donburi.Entry, _ *rendering.View) {
			sprite.Component.Get(e).SetFrame(1)
		}, 6, 4, white},
		{"view matrix", func(_ *donburi.Entry, view *rendering.View) {
			view.Matrix.Translate(2, 0)
		}, 6, 4, white},
		{"texture not loaded", func(e *donburi.Entry, _ *rendering.View) {
			sprite.Component.Get(e).SetTexture(0)
		}, 4, 4, blank},
		{"no transform", func(e *donburi.Entry, _ *rendering.View) {
			e.RemoveComponent(transform.Component)
		}, 0, 0, blank},
	}

	got := make([]color.Color, len(tests))
	ebitentest.InFrame(t, func() {
		square := ebiten.NewImage(2, 2)
		square.Fill(white)
		tex := images.NewTexture(square)
		// frame 1 is the same square trimmed 2px in from the left of a 4x2 frame
		if err := tex.SliceRegions([]images.Region{
			{Rect: image.Rect(0, 0, 2, 2), Size: image.Pt(2, 2)},
			{Rect: image.Rect(0, 0, 2, 2), Offset: image.Pt(2, 0), Size: image.Pt(4, 2)},
		}); err != nil {
			t.Fatalf("SliceRegions() = %v, want nil", err)
		}
		p := rendering.NewPipeline(sprite.RenderAdapter(textures{tex}))

		for i, tt := range tests {
			w := donburi.NewWorld()
			e := w.Entry(w.Create(rendering.Component, sprite.Component, transform.Component))
			sprite.Component.Get(e).SetTexture(1)
			transform.Component.Get(e).SetPosition(4, 4)
			var view rendering.View
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
