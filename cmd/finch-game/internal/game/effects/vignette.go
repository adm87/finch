package effects

import (
	_ "embed"

	"github.com/hajimehoshi/ebiten/v2"

	"github.com/adm87/finch/pkg/engine"
)

//go:embed vignette.kage
var vignetteSrc []byte

// Vignette darkens the corners of the frame.
type Vignette struct {
	shader *ebiten.Shader
	frame  *ebiten.Image
	op     ebiten.DrawRectShaderOptions
}

func NewVignette(strength float64) (*Vignette, error) {
	shader, err := ebiten.NewShader(vignetteSrc)
	if err != nil {
		return nil, err
	}
	v := &Vignette{shader: shader}
	v.op.Uniforms = map[string]any{"Strength": float32(strength)}
	return v, nil
}

func (v *Vignette) Pass() engine.DrawPass {
	return func(inner engine.Drawer) engine.Drawer {
		return engine.DrawerFunc(func(target *ebiten.Image) {
			size := target.Bounds().Size()
			if v.frame == nil || v.frame.Bounds().Size() != size {
				if v.frame != nil {
					v.frame.Deallocate()
				}
				v.frame = ebiten.NewImage(size.X, size.Y)
			} else {
				v.frame.Clear()
			}

			inner.Draw(v.frame)

			v.op.Images[0] = v.frame
			target.DrawRectShader(size.X, size.Y, v.shader, &v.op)
		})
	}
}
