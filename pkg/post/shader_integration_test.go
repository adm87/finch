//go:build integration

package post_test

import (
	"image/color"
	"testing"

	"github.com/hajimehoshi/ebiten/v2"

	"github.com/adm87/finch/pkg/engine"
	"github.com/adm87/finch/pkg/post"
)

var scale = []byte(`//kage:unit pixels

package main

var Amount float

func Fragment(dst vec4, src vec2, color vec4) vec4 {
	return imageSrc0At(src) * Amount
}
`)

func TestIntegrationShaderAppliesToFrame(t *testing.T) {
	shader, err := ebiten.NewShader(scale)
	if err != nil {
		t.Fatalf("NewShader() = %v", err)
	}

	white := engine.DrawerFunc(func(target *ebiten.Image) { target.Fill(color.White) })
	drawer := engine.DrawPipeline(white, post.Shader(shader, map[string]any{"Amount": float32(0.5)}))
	target := ebiten.NewImage(8, 8)

	var got color.RGBA
	var drawn bool
	err = engine.Run(
		engine.WithUpdater(engine.UpdaterFunc(func(float64) error {
			if drawn {
				return ebiten.Termination
			}
			return nil
		})),
		engine.WithDrawer(engine.DrawerFunc(func(*ebiten.Image) {
			if drawn {
				return
			}
			drawer.Draw(target)
			got = color.RGBAModel.Convert(target.At(4, 4)).(color.RGBA)
			drawn = true
		})),
	)
	if err != nil {
		t.Fatalf("Run() = %v", err)
	}
	if got.A < 126 || got.A > 129 {
		t.Errorf("pixel = %v, want about half white", got)
	}
}
