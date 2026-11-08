package engine

import "github.com/hajimehoshi/ebiten/v2"

// Drawer draws a frame.
type Drawer interface {
	Draw(target *ebiten.Image)
}

// DrawFunc adapts a function to a Drawer.
type DrawFunc func(target *ebiten.Image)

func (f DrawFunc) Draw(target *ebiten.Image) {
	f(target)
}

// DrawPass wraps a Drawer to draw before, after, or around it.
type DrawPass func(Drawer) Drawer

// DrawPipeline wraps drawer in each pass in order.
func DrawPipeline(drawer Drawer, passes ...DrawPass) Drawer {
	for _, pass := range passes {
		drawer = pass(drawer)
	}
	return drawer
}

func noopDraw(_ *ebiten.Image) {}
