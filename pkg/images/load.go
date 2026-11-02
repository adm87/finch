package images

import (
	"bytes"
	"errors"
	"image"
	_ "image/png" // registers the .png decoder

	"github.com/hajimehoshi/ebiten/v2"

	"github.com/adm87/finch/pkg/resource"
)

// Decode decodes an image file. It doesn't touch ebiten, so it can run off
// the game loop.
func Decode(raw []byte) (image.Image, error) {
	img, _, err := image.Decode(bytes.NewReader(raw))
	return img, err
}

// Load reads the images from src that cache doesn't have yet and stores them
// under resource.Key(src, name). Images that fail are reported and skipped.
func Load(res *resource.Resources, cache *Cache, src resource.Source, names ...string) error {
	l := resource.NewLoader(res, Decode)
	for _, name := range names {
		if _, ok := cache.Lookup(resource.Key(src, name)); !ok {
			l.Queue(src, name)
		}
	}

	results := l.Load()
	errList := []error{resource.Err(results)}
	for _, r := range results {
		if r.Err != nil {
			continue
		}
		img := ebiten.NewImageFromImage(r.Value)
		if _, err := cache.Put(resource.Key(r.Source, r.Name), img); err != nil {
			img.Deallocate()
			errList = append(errList, err)
		}
	}
	return errors.Join(errList...)
}
