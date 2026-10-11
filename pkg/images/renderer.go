package images

import (
	"image/color"

	"github.com/adm87/finch/pkg/ecs/components/transform"
	"github.com/adm87/finch/pkg/ecs/rendering"
	"github.com/adm87/finch/pkg/geom"
	"github.com/hajimehoshi/ebiten/v2"
)

type ImageRenderer struct {
	missing     *ebiten.Image
	drawOptions *ebiten.DrawImageOptions
}

func NewImageRenderer() *ImageRenderer {
	missing := ebiten.NewImage(10, 10)
	missing.Fill(color.RGBA{R: 255, B: 255, A: 255})
	return &ImageRenderer{
		missing:     missing,
		drawOptions: &ebiten.DrawImageOptions{},
	}
}

func (ir *ImageRenderer) Render(target *ebiten.Image, rc *rendering.RenderCandidate, viewport geom.Rect, matrix ebiten.GeoM) {
	ir.drawOptions.GeoM = transform.GetGeoMOf(rc.Transform, rc.Matrix)
	ir.drawOptions.GeoM.Concat(matrix)
	target.DrawImage(ir.missing, ir.drawOptions)
}
