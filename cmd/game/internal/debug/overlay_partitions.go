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

// Partitions outlines the occupied partition cells. Bigger cells get thicker lines.
type Partitions struct {
	World *world.World
}

func (o *Partitions) Draw(target *ebiten.Image, ctx *engine.DrawContext) error {
	if ctx.View.Area() == 0 {
		return nil
	}
	for cell := range o.World.PartitionCells(ctx.View) {
		clr := partitionColors[cell.Partition%len(partitionColors)]
		strokeWorldRect(target, ctx.Matrix, cell.Bounds, float32(1+cell.Partition), clr)
	}
	return nil
}
