package game

import (
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
)

type Updater struct{}

func (u *Updater) Update(dt float64) error {
	if inpututil.IsKeyJustPressed(ebiten.KeyEscape) {
		return ebiten.Termination
	}
	return nil
}

func (u *Updater) FixedUpdate(dt float64) error {
	return nil
}

func (u *Updater) LaterUpdate(dt float64) error {
	return nil
}
