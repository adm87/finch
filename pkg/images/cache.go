package images

import (
	"bytes"
	_ "image/jpeg" // registers the JPEG decoder
	_ "image/png"  // registers the PNG decoder
	"io/fs"

	"github.com/hajimehoshi/ebiten/v2/ebitenutil"

	"github.com/adm87/finch/pkg/assets"
)

// Register makes the image cache and routes image extensions on router to it.
func Register(fsys fs.FS, router *assets.Router) (*assets.Cache[*Texture], error) {
	cache := assets.NewCache(fsys,
		assets.ImporterFunc[*Texture](decode),
		assets.ReleaserFunc[*Texture](release),
	)
	if err := assets.Route(router, cache, ".png", ".jpeg", ".jpg"); err != nil {
		return nil, err
	}
	return cache, nil
}

// decode reads the file at path and decodes it into a Texture.
// The texture is unsliced, so it has one frame that is the whole image.
func decode(fsys fs.FS, path string) (*Texture, error) {
	data, err := fs.ReadFile(fsys, path)
	if err != nil {
		return nil, err
	}

	img, _, err := ebitenutil.NewImageFromReader(bytes.NewReader(data))
	if err != nil {
		return nil, err
	}

	return NewTexture(img), nil
}

func release(texture *Texture) {
	texture.source.Deallocate()
	texture.source = nil
	texture.frames = nil
}
