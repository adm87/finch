package debug

import (
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"

	"github.com/adm87/finch/pkg/application"
)

var partitionColors = []color.RGBA{
	{R: 255, G: 90, B: 90, A: 200},
	{R: 90, G: 220, B: 90, A: 200},
	{R: 90, G: 150, B: 255, A: 200},
}

// OverlayPartitions outlines the occupied partition cells in view. Larger cells
// get thicker lines so overlapping partitions stay distinguishable.
func OverlayPartitions(target *ebiten.Image, app *application.Application) error {
	region, viewGeoM := app.View().Viewport()
	if region.Area() == 0 {
		return nil
	}

	toScreen := worldToScreen(app.Screen(), viewGeoM)
	for cell := range app.World().PartitionCells(region) {
		clr := partitionColors[cell.Partition%len(partitionColors)]
		strokeWorldRect(target, toScreen, cell.Bounds, float32(1+cell.Partition), clr)
	}
	return nil
}
