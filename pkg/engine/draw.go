package engine

import "github.com/hajimehoshi/ebiten/v2"

// Drawer is an interface that defines a type that can draw to an Ebitengine screen.
type Drawer interface {
	Draw(screen *ebiten.Image)
}

// DrawerFunc is a function type that implements the Drawer interface.
// It allows a function to be used as a Drawer by defining a Draw method that calls the function.
type DrawerFunc func(screen *ebiten.Image)

func (f DrawerFunc) Draw(screen *ebiten.Image) {
	f(screen)
}

func noopDraw(_ *ebiten.Image) {
	// noop draw
}
