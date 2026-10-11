package engine

import "github.com/hajimehoshi/ebiten/v2"

type GameUpdate interface {
	Update() error
}

type GameDraw interface {
	Draw(target *ebiten.Image) error
}

type GameLayout interface {
	Layout(outsideWidth, outsideHeight int) (screenWidth, screenHeight int)
}

type noopUpdate struct{}

func (n *noopUpdate) Update() error {
	return nil
}

type noopDraw struct{}

func (n *noopDraw) Draw(target *ebiten.Image) error {
	return nil
}

type noopLayout struct{}

func (n *noopLayout) Layout(outsideWidth, outsideHeight int) (screenWidth, screenHeight int) {
	return outsideWidth, outsideHeight
}
