package sprite_test

import (
	"github.com/adm87/finch/pkg/assets"
	"github.com/adm87/finch/pkg/images"
)

// textures serves its texture for handle 1 and nothing for any other.
type textures struct {
	texture *images.Texture
}

func (t textures) Get(h assets.Handle[*images.Texture]) (*images.Texture, bool) {
	return t.texture, h == 1
}
