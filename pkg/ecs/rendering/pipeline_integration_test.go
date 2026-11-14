//go:build integration

package rendering_test

import (
	"image/color"
	"testing"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/yohamta/donburi"

	"github.com/adm87/finch/pkg/ecs/rendering"
	"github.com/adm87/finch/pkg/internal/ebitentest"
)

var (
	red   = color.RGBA{R: 255, A: 255}
	green = color.RGBA{G: 255, A: 255}
)

type draw struct {
	color        color.RGBA
	layer, depth int16
}

// Every draw covers the same pixel, so the one drawn last is the one left on top.
func TestIntegrationDrawOrder(t *testing.T) {
	tests := []struct {
		name  string
		draws []draw
	}{
		{"layer beats depth", []draw{{green, 1, -100}, {red, 0, 100}}},
		{"negative layer under", []draw{{green, 0, 0}, {red, -1, 0}}},
		{"depth within layer", []draw{{green, 0, 5}, {red, 0, -5}}},
		{"negative depth under", []draw{{green, 0, 0}, {red, 0, -1}}},
		{"push order breaks ties", []draw{{red, 0, 0}, {green, 0, 0}}},
	}

	got := make([]color.Color, len(tests))
	ebitentest.InFrame(t, func() {
		for i, tt := range tests {
			p := rendering.NewPipeline(rendering.AdapterFunc(func(q *rendering.Queue, _ []*donburi.Entry, _ rendering.View) {
				for _, d := range tt.draws {
					img := ebiten.NewImage(1, 1)
					img.Fill(d.color)
					q.Push(rendering.Command{Image: img, Layer: d.layer, Depth: d.depth})
				}
			}))
			target := ebiten.NewImage(1, 1)
			p.Draw(target, rendering.View{}, nil)
			got[i] = target.At(0, 0)
		}
	})

	for i, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got[i] != green {
				t.Errorf("top pixel = %v, want %v", got[i], green)
			}
		})
	}
}
