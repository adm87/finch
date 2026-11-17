package sprite_test

import (
	"github.com/hajimehoshi/ebiten/v2"

	"github.com/adm87/finch/pkg/assets"
)

// images serves its image for handle 1 and nothing for any other.
type images struct {
	square *ebiten.Image
}

func (i images) Get(h assets.Handle[*ebiten.Image]) (*ebiten.Image, bool) {
	return i.square, h == 1
}
