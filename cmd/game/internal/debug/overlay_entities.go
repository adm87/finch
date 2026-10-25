package debug

import (
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"

	"github.com/adm87/finch/pkg/engine"
	"github.com/adm87/finch/pkg/transform"
	"github.com/adm87/finch/pkg/world"
)

var (
	entityBoundsColor = color.RGBA{R: 255, G: 220, B: 60, A: 230}
	entityOriginColor = color.RGBA{R: 255, G: 255, B: 255, A: 230}
)

const originMarkerSize = 3

// OverlayEntityBounds outlines each entity's bounds and marks its origin.
func OverlayEntityBounds(w *world.World) engine.Drawer {
	return engine.DrawerFunc(func(target *ebiten.Image, ctx *engine.DrawContext) error {
		if !GetDebug(w.ECS()).HasOverlay(OverlayTypeEntities) {
			return nil
		}
		for _, entry := range ctx.Entries {
			t, tOk := transform.GetTransform(entry)
			b, bOk := transform.GetBounds(entry)
			if !tOk || !bOk {
				continue
			}

			// The entries are a rough match, so skip anything outside the view.
			rect := b.WorldRect(t)
			if !rect.Intersects(ctx.View) {
				continue
			}
			strokeWorldRect(target, ctx.Matrix, rect, 1, entityBoundsColor)

			px, py := t.Position()
			x, y := applyFloat32(ctx.Matrix, px, py)
			vector.StrokeLine(target, x-originMarkerSize, y, x+originMarkerSize, y, 1, entityOriginColor, false)
			vector.StrokeLine(target, x, y-originMarkerSize, x, y+originMarkerSize, 1, entityOriginColor, false)
		}
		return nil
	})
}
