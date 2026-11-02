package engine

// Display is an interface that defines a type that can handle the layout of the game window.
type Display interface {
	Layout(int, int) (int, int)
}

// DisplayFunc is a function type that implements the Display interface.
// It allows a function to be used as a Display by defining a Layout method that calls the function.
type DisplayFunc func(int, int) (int, int)

func (f DisplayFunc) Layout(outsideWidth, outsideHeight int) (int, int) {
	return f(outsideWidth, outsideHeight)
}

func noopLayout(outsideWidth, outsideHeight int) (int, int) {
	return outsideWidth, outsideHeight
}
