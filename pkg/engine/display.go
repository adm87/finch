package engine

// Display sizes the screen for the window.
type Display interface {
	Layout(outsideWidth, outsideHeight float64) (layoutWidth, layoutHeight float64)
}

// DisplayFunc adapts a function to a Display.
type DisplayFunc func(outsideWidth, outsideHeight float64) (layoutWidth, layoutHeight float64)

func (f DisplayFunc) Layout(outsideWidth, outsideHeight float64) (layoutWidth, layoutHeight float64) {
	return f(outsideWidth, outsideHeight)
}

func noopLayout(outsideWidth, outsideHeight float64) (layoutWidth, layoutHeight float64) {
	return outsideWidth, outsideHeight
}
