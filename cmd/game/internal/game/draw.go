package game

import (
	"fmt"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"

	"github.com/adm87/finch/pkg/display"
	"github.com/adm87/finch/pkg/engine"
	"github.com/adm87/finch/pkg/hexcolor"
)

type Drawer struct {
	drawer     engine.GameDraw
	screen     *display.Screen
	view       *display.View
	clearColor hexcolor.Color
}

func NewDrawer(drawer engine.GameDraw, screen *display.Screen, view *display.View, options ...DrawOption) *Drawer {
	opts := DefaultDrawOptions()
	for _, o := range options {
		o(opts)
	}

	return &Drawer{
		drawer:     drawer,
		screen:     screen,
		view:       view,
		clearColor: opts.ClearColor,
	}
}

func (d *Drawer) Draw(target *ebiten.Image) error {
	buf := d.screen.Buffer()
	buf.Fill(d.clearColor.ToRGBA())

	if err := d.drawer.Draw(buf); err != nil {
		return err
	}

	target.DrawImage(buf, d.screen.Options())

	x, y := d.view.WorldToScreen(0, 0)
	ebitenutil.DebugPrintAt(target, "0, 0", int(x), int(y))

	fps := fmt.Sprintf("FPS: %.2f", ebiten.ActualFPS())
	ebitenutil.DebugPrintAt(target, fps, 10, 10)

	return nil
}
