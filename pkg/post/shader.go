package post

import (
	"github.com/hajimehoshi/ebiten/v2"

	"github.com/adm87/finch/pkg/engine"
)

// Shader draws the frame into an offscreen image, then onto the target through
// shader. The frame is the shader's first source image. uniforms is read on
// every draw, so changes to it show on the next frame.
func Shader(shader *ebiten.Shader, uniforms map[string]any) engine.DrawPass {
	if shader == nil {
		panic("post: nil shader")
	}
	return func(inner engine.Drawer) engine.Drawer {
		var frame *ebiten.Image
		op := &ebiten.DrawRectShaderOptions{Uniforms: uniforms}

		return engine.DrawerFunc(func(target *ebiten.Image) {
			size := target.Bounds().Size()
			if frame == nil || frame.Bounds().Size() != size {
				if frame != nil {
					frame.Deallocate()
				}
				frame = ebiten.NewImage(size.X, size.Y)
			} else {
				frame.Clear()
			}

			inner.Draw(frame)

			op.Images[0] = frame
			target.DrawRectShader(size.X, size.Y, shader, op)
		})
	}
}
