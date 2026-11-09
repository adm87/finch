package effects

import (
	"github.com/hajimehoshi/ebiten/v2"

	"github.com/adm87/finch/pkg/engine"
)

// pass draws everything before it into its own image, then runs that image
// through a shader onto the target. When it's off, it stays out of the way.
type pass struct {
	shader  *ebiten.Shader
	frame   *ebiten.Image
	op      ebiten.DrawRectShaderOptions
	enabled bool
	used    bool
}

func newPass(src []byte, uniforms map[string]any) (pass, error) {
	shader, err := ebiten.NewShader(src)
	if err != nil {
		return pass{}, err
	}
	p := pass{shader: shader, enabled: true}
	p.op.Uniforms = uniforms
	return p, nil
}

func (p *pass) Toggle() {
	p.enabled = !p.enabled
}

// Pass can only be used once, since the effect has just one image to draw into.
func (p *pass) Pass() engine.DrawPass {
	if p.used {
		panic("effects: pass already in a pipeline")
	}
	p.used = true

	return func(inner engine.Drawer) engine.Drawer {
		return engine.DrawFunc(func(target *ebiten.Image) {
			if !p.enabled {
				inner.Draw(target)
				return
			}

			size := target.Bounds().Size()
			if p.frame == nil || p.frame.Bounds().Size() != size {
				if p.frame != nil {
					p.frame.Deallocate()
				}
				p.frame = ebiten.NewImage(size.X, size.Y)
			} else {
				p.frame.Clear()
			}

			inner.Draw(p.frame)

			p.op.Images[0] = p.frame
			target.DrawRectShader(size.X, size.Y, p.shader, &p.op)
		})
	}
}
