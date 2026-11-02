package engine

// Updater is an interface that defines a type that can update the game state.
type Updater interface {
	Update(dt float64) error
}

// UpdaterFunc is a function type that implements the Updater interface.
// It allows a function to be used as an Updater by defining an Update method that calls the function.
type UpdaterFunc func(dt float64) error

func (f UpdaterFunc) Update(dt float64) error {
	return f(dt)
}

func noopUpdate(_ float64) error {
	return nil
}
