package game

import (
	"fmt"
	"image/color"

	"github.com/adm87/finch/pkg/ecs/components/transform"
	"github.com/adm87/finch/pkg/engine"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
)

type GameDraw struct {
	drawer engine.GameDraw
	screen *Screen
	view   *View
}

func NewGameDraw(drawer engine.GameDraw, screen *Screen, view *View) *GameDraw {
	return &GameDraw{
		drawer: drawer,
		screen: screen,
		view:   view,
	}
}

func (gd *GameDraw) Draw(target *ebiten.Image) error {
	gd.screen.buffer.Fill(color.RGBA{R: 100, G: 149, B: 237, A: 255})

	if err := gd.drawer.Draw(gd.screen.buffer); err != nil {
		return err
	}

	x, y := gd.view.WorldToScreen(0, 0)
	ebitenutil.DebugPrintAt(gd.screen.buffer, "0, 0 (world)", int(x), int(y))

	camBounds := transform.GetRect(gd.view.camera)

	x, y = gd.view.WorldToScreen(camBounds.Center())
	ebitenutil.DebugPrintAt(gd.screen.buffer, "0, 0 (safe area center)", int(x), int(y))

	target.DrawImage(gd.screen.buffer, gd.screen.options)

	cx, cy := ebiten.CursorPositionF()
	wx, wy := gd.view.ScreenToWorld(cx, cy)
	ebitenutil.DebugPrintAt(target, fmt.Sprintf("cursor world: %.0f, %.0f", wx, wy), 10, 30)

	safeArea := gd.screen.SafeArea()
	ebitenutil.DebugPrintAt(target, fmt.Sprintf("safe area: %v", safeArea), 10, 10)

	return nil
}
