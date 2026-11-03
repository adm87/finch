package effects

import (
	_ "embed"

	"github.com/hajimehoshi/ebiten/v2"
)

//go:embed vignette.kage
var vignette []byte

func Vignette() (*ebiten.Shader, error) {
	return ebiten.NewShader(vignette)
}
