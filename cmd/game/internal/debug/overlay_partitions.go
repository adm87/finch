package debug

import (
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"

	"github.com/adm87/finch/pkg/engine"
	"github.com/adm87/finch/pkg/world"
)

var partitionColors = []color.RGBA{
	{R: 255, G: 90, B: 90, A: 200},
	{R: 90, G: 220, B: 90, A: 200},
	{R: 90, G: 150, B: 255, A: 200},
}

func OverlayPartitions(w *world.World) engine.Drawer {
	return engine.DrawerFunc(func(target *ebiten.Image, ctx *engine.DrawContext) error {
		if !GetDebug(w.ECS()).HasOverlay(OverlayTypePartitions) {
			return nil
		}
		if ctx.View.Area() == 0 {
			return nil
		}
		for cell := range w.PartitionCells(ctx.View) {
			clr := partitionColors[cell.Partition%len(partitionColors)]
			strokeWorldRect(target, ctx.Matrix, cell.Bounds, float32(1+cell.Partition), clr)
		}
		return nil
	})
}
