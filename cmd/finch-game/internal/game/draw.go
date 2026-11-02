package game

import (
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
)

type Drawer struct{}

func (d *Drawer) Draw(screen *ebiten.Image) {
	ebitenutil.DebugPrint(screen, "Hello, Finch!")
}
