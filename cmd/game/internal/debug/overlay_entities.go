package debug

import (
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
	"github.com/yohamta/donburi"

	"github.com/adm87/finch/pkg/application"
	"github.com/adm87/finch/pkg/transform"
)

var (
	entityBoundsColor = color.RGBA{R: 255, G: 220, B: 60, A: 230}
	entityOriginColor = color.RGBA{R: 255, G: 255, B: 255, A: 230}
)

const originMarkerSize = 3

// EntityBounds outlines the bounds of each entity in view and marks its origin.
type EntityBounds struct {
	entries []*donburi.Entry
}

func (o *EntityBounds) Draw(target *ebiten.Image, app *application.Application) error {
	region, viewGeoM := app.View().Viewport()
	if region.Area() == 0 {
		return nil
	}

	toScreen := worldToScreen(app.Screen(), viewGeoM)
	o.entries = app.World().Query(region, o.entries[:0])

	for _, entry := range o.entries {
		t, tOk := transform.GetTransform(entry)
		b, bOk := transform.GetBounds(entry)
		if !tOk || !bOk {
			continue
		}

		// The world query is a broad phase; keep only bounds that really overlap the view.
		rect := b.WorldRect(t)
		if !rect.Intersects(region) {
			continue
		}
		strokeWorldRect(target, toScreen, rect, 1, entityBoundsColor)

		x, y := toScreen(t.Position())
		vector.StrokeLine(target, x-originMarkerSize, y, x+originMarkerSize, y, 1, entityOriginColor, false)
		vector.StrokeLine(target, x, y-originMarkerSize, x, y+originMarkerSize, 1, entityOriginColor, false)
	}
	return nil
}
