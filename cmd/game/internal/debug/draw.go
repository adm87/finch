package debug

import (
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"

	"github.com/adm87/finch/pkg/display"
	"github.com/adm87/finch/pkg/geom"
)

type screenPoint func(x, y float64) (float32, float32)

// worldToScreen maps world coordinates to window coordinates, like
// View.WorldToScreen, using a view matrix read once per frame.
func worldToScreen(screen *display.Screen, viewGeoM ebiten.GeoM) screenPoint {
	return func(x, y float64) (float32, float32) {
		sx, sy := screen.SafeAreaToScreen(viewGeoM.Apply(x, y))
		return float32(sx), float32(sy)
	}
}

// strokeWorldRect outlines rect through toScreen, so it stays correct under
// camera rotation and zoom.
func strokeWorldRect(target *ebiten.Image, toScreen screenPoint, rect geom.Rect, width float32, clr color.Color) {
	minX, minY := rect.Min()
	maxX, maxY := rect.Max()
	x0, y0 := toScreen(minX, minY)
	x1, y1 := toScreen(maxX, minY)
	x2, y2 := toScreen(maxX, maxY)
	x3, y3 := toScreen(minX, maxY)

	vector.StrokeLine(target, x0, y0, x1, y1, width, clr, false)
	vector.StrokeLine(target, x1, y1, x2, y2, width, clr, false)
	vector.StrokeLine(target, x2, y2, x3, y3, width, clr, false)
	vector.StrokeLine(target, x3, y3, x0, y0, width, clr, false)
}
