package effects

import (
	_ "embed"

	"github.com/hajimehoshi/ebiten/v2"

	"github.com/adm87/finch/pkg/engine"
	"github.com/adm87/finch/pkg/post"
)

//go:embed crt.kage
var crt []byte

// CRT is an animated old-monitor effect. Run it as an updater so it keeps time.
type CRT struct {
	shader   *ebiten.Shader
	uniforms map[string]any
	time     float64
}

func NewCRT(lines float64) (*CRT, error) {
	shader, err := ebiten.NewShader(crt)
	if err != nil {
		return nil, err
	}
	return &CRT{
		shader:   shader,
		uniforms: map[string]any{"Lines": float32(lines), "Time": float32(0)},
	}, nil
}

func (c *CRT) Update(dt float64) error {
	c.time += dt
	c.uniforms["Time"] = float32(c.time)
	return nil
}

func (c *CRT) Pass() engine.DrawPass {
	return post.Shader(c.shader, c.uniforms)
}
