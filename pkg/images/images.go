package images

import (
	"bytes"
	_ "image/jpeg" // registers the JPEG decoder
	_ "image/png"  // registers the PNG decoder
	"io/fs"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"

	"github.com/adm87/finch/pkg/assets"
)

// Register makes the image cache and routes image extensions on router to it.
func Register(fsys fs.FS, router *assets.Router) (*assets.Cache[*ebiten.Image], error) {
	cache := assets.NewCache(fsys,
		assets.ImporterFunc[*ebiten.Image](decode),
		assets.ReleaserFunc[*ebiten.Image](release),
	)
	if err := assets.Route(router, cache, ".png", ".jpeg", ".jpg"); err != nil {
		return nil, err
	}
	return cache, nil
}

// decode reads the PNG or JPEG image at path.
func decode(fsys fs.FS, path string) (*ebiten.Image, error) {
	data, err := fs.ReadFile(fsys, path)
	if err != nil {
		return nil, err
	}

	img, _, err := ebitenutil.NewImageFromReader(bytes.NewReader(data))
	return img, err
}

func release(img *ebiten.Image) {
	img.Deallocate()
}
