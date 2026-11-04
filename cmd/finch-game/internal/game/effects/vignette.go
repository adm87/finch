package effects

import (
	_ "embed"
	"fmt"
)

//go:embed vignette.kage
var vignetteSrc []byte

// Vignette darkens the corners of the frame.
type Vignette struct {
	pass
}

func NewVignette(strength float64) (*Vignette, error) {
	p, err := newPass(vignetteSrc, map[string]any{"Strength": float32(strength)})
	if err != nil {
		return nil, fmt.Errorf("vignette shader: %w", err)
	}
	return &Vignette{pass: p}, nil
}
