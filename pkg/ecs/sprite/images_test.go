package sprite_test

import "github.com/hajimehoshi/ebiten/v2"

// images serves its image for handle 1 and nothing for any other.
type images struct {
	square *ebiten.Image
}

func (i images) Get(handle uint64) *ebiten.Image {
	if handle != 1 {
		return nil
	}
	return i.square
}
