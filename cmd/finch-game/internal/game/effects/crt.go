package effects

import (
	_ "embed"
	"fmt"
)

//go:embed crt.kage
var crtSrc []byte

// CRT is an animated old-monitor effect. Run it as an updater so it keeps time.
type CRT struct {
	pass
	time float64
}

func NewCRT(lines float64) (*CRT, error) {
	p, err := newPass(crtSrc, map[string]any{"Lines": float32(lines), "Time": float32(0)})
	if err != nil {
		return nil, fmt.Errorf("crt shader: %w", err)
	}
	return &CRT{pass: p}, nil
}

func (c *CRT) Update(dt float64) error {
	c.time += dt
	c.op.Uniforms["Time"] = float32(c.time)
	return nil
}
