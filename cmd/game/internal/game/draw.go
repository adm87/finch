package game

import (
	"fmt"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"

	"github.com/adm87/finch/pkg/display"
	"github.com/adm87/finch/pkg/engine"
	"github.com/adm87/finch/pkg/hexcolor"
)

type GameDraw struct {
	drawer     engine.GameDraw
	screen     *display.Screen
	view       *display.View
	clearColor hexcolor.Color
}

func NewGameDraw(drawer engine.GameDraw, screen *display.Screen, view *display.View, options ...DrawOption) *GameDraw {
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
	buf := gd.screen.Buffer()
	buf.Fill(gd.clearColor.ToRGBA())

	if err := gd.drawer.Draw(buf); err != nil {
		return err
	}

	target.DrawImage(buf, gd.screen.Options())

	x, y := gd.view.WorldToScreen(0, 0)
	ebitenutil.DebugPrintAt(target, "0, 0", int(x), int(y))

	fps := fmt.Sprintf("FPS: %.2f", ebiten.ActualFPS())
	ebitenutil.DebugPrintAt(target, fps, 10, 10)

	return nil
}
