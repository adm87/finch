package game

import (
	"fmt"
	"image/color"

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

	target.DrawImage(gd.screen.buffer, gd.screen.options)

	fps := fmt.Sprintf("FPS: %.2f", ebiten.ActualFPS())
	ebitenutil.DebugPrintAt(target, fps, 10, 10)

	return nil
}
