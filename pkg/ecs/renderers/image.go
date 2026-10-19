package renderers

import (
	"image/color"

	"github.com/adm87/finch/pkg/ecs/components"
	"github.com/adm87/finch/pkg/ecs/rendering"
	"github.com/adm87/finch/pkg/geom"
	"github.com/adm87/finch/pkg/storage"
	"github.com/hajimehoshi/ebiten/v2"
)

type ImageRenderer struct {
	missing     *ebiten.Image
	drawOptions *ebiten.DrawImageOptions
	cache       *storage.ImageCache
}

func NewImageRenderer(cache *storage.ImageCache) *ImageRenderer {
	missing := ebiten.NewImage(10, 10)
	missing.Fill(color.RGBA{R: 255, B: 255, A: 255})
	return &ImageRenderer{
		missing:     missing,
		drawOptions: &ebiten.DrawImageOptions{},
		cache:       cache,
	}
}

func (ir *ImageRenderer) Render(target *ebiten.Image, rc *rendering.RenderCandidate, viewport geom.Rect, matrix ebiten.GeoM) {
	src := ir.missing
	ax, ay := 0.0, 0.0

	ir.drawOptions.ColorScale.Reset()

	if img, iOk := components.GetImage(rc.Entry); iOk {
		ax, ay = img.Anchor()
		frame := img.Frame()

		ir.drawOptions.ColorScale.ScaleWithColor(img.Color())

		if cached, _ := ir.cache.GetFrame(img.Handle(), frame); cached != nil {
			src = cached
		}
	}

	t := rc.Transform
	m := rc.Matrix

	ir.drawOptions.GeoM.Reset()
	ir.drawOptions.GeoM.Translate(
		-float64(src.Bounds().Dx())*ax,
		-float64(src.Bounds().Dy())*ay,
	)
	ir.drawOptions.GeoM.Concat(m.GeoM(t))
	ir.drawOptions.GeoM.Concat(matrix)

	target.DrawImage(src, ir.drawOptions)
}
