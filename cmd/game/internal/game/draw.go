package game

import (
	"image/color"

	"github.com/adm87/finch/pkg/ecs/components/transform"
	"github.com/adm87/finch/pkg/engine"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
)

type GameDraw struct {
	drawer     engine.GameDraw
	screen     *Screen
	view       *View
	clearColor color.RGBA
}

func NewGameDraw(drawer engine.GameDraw, screen *Screen, view *View, options ...DrawOption) *GameDraw {
	opts := DefaultDrawOptions()
	for _, o := range options {
		o(opts)
	}

	return &GameDraw{
		drawer:     drawer,
		screen:     screen,
		view:       view,
		clearColor: opts.ClearColor,
	}
}

func (gd *GameDraw) Draw(target *ebiten.Image) error {
	gd.screen.buffer.Fill(gd.clearColor)

	if err := gd.drawer.Draw(gd.screen.buffer); err != nil {
		return err
	}

	x, y := gd.view.WorldToScreen(0, 0)
	ebitenutil.DebugPrintAt(gd.screen.buffer, "0, 0", int(x), int(y))

	camBounds := transform.GetRect(gd.view.camera)

	cx, cy := gd.view.WorldToScreen(camBounds.Center())
	ebitenutil.DebugPrintAt(gd.screen.buffer, "center", int(cx)-17, int(cy)-7)

	left, top := camBounds.Min()
	right, bottom := camBounds.Max()

	x, y = gd.view.WorldToScreen(left, 0)
	ebitenutil.DebugPrintAt(gd.screen.buffer, "left", int(x), int(cy))

	x, y = gd.view.WorldToScreen(right, 0)
	ebitenutil.DebugPrintAt(gd.screen.buffer, "right", int(x)-35, int(cy))

	x, y = gd.view.WorldToScreen(0, top)
	ebitenutil.DebugPrintAt(gd.screen.buffer, "top", int(cx), int(y))

	x, y = gd.view.WorldToScreen(0, bottom)
	ebitenutil.DebugPrintAt(gd.screen.buffer, "bottom", int(cx), int(y)-15)

	target.DrawImage(gd.screen.buffer, gd.screen.options)

	return nil
}
