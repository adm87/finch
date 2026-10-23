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

func (ir *Renderer) Render(target *ebiten.Image, rc *rendering.Candidate, viewport geom.Rect, matrix ebiten.GeoM) {
	src := ir.missing
	ax, ay := 0.0, 0.0

	ir.drawOptions.ColorScale.Reset()

	if img, iOk := GetImage(rc.Entry); iOk {
		ax, ay = img.Anchor()
		frame := img.Frame()

		ir.drawOptions.ColorScale.ScaleWithColor(img.Color())

		if cached, _ := ir.cache.GetFrame(img.Handle(), frame); cached != nil {
			src = cached
		}
	}

	ir.drawOptions.GeoM.Reset()
	ir.drawOptions.GeoM.Translate(
		-float64(src.Bounds().Dx())*ax,
		-float64(src.Bounds().Dy())*ay,
	)
	ir.drawOptions.GeoM.Concat(rc.Transform.GeoM())
	ir.drawOptions.GeoM.Concat(matrix)

	target.DrawImage(src, ir.drawOptions)
}
