package images

import (
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"

	"github.com/adm87/finch/pkg/geom"
	"github.com/adm87/finch/pkg/rendering"
)

type Renderer struct {
	missing     *ebiten.Image
	drawOptions *ebiten.DrawImageOptions
	cache       *Cache
}

func NewRenderer(cache *Cache) *Renderer {
	missing := ebiten.NewImage(10, 10)
	missing.Fill(color.RGBA{R: 255, B: 255, A: 255})
	return &Renderer{
		missing:     missing,
		drawOptions: &ebiten.DrawImageOptions{},
		cache:       cache,
	}
}

func (ir *Renderer) Render(target *ebiten.Image, rc *rendering.Candidate, _ geom.Rect, matrix geom.Matrix) {
	frame := untrimmed(ir.missing)
	ax, ay := 0.0, 0.0

	ir.drawOptions.ColorScale.Reset()

	if img, iOk := GetImage(rc.Entry); iOk {
		ax, ay = img.Anchor()

		ir.drawOptions.ColorScale.ScaleWithColor(img.Color())

		if cached, err := ir.cache.Frame(img.Handle(), img.Frame()); err == nil {
			frame = cached
		}
	}

	ir.drawOptions.GeoM.Reset()
	ir.drawOptions.GeoM.Translate(
		float64(frame.Offset.X)-float64(frame.Size.X)*ax,
		float64(frame.Offset.Y)-float64(frame.Size.Y)*ay,
	)
	ir.drawOptions.GeoM.Concat(rc.Transform.GeoM())
	ir.drawOptions.GeoM.Concat(matrix)

	target.DrawImage(frame.Image, ir.drawOptions)
}
