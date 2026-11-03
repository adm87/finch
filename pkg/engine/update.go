package engine

// Updater advances the game state by dt seconds.
type Updater interface {
	Update(dt float64) error
}

// UpdaterFunc adapts a function to an Updater.
type UpdaterFunc func(dt float64) error

func (f UpdaterFunc) Update(dt float64) error {
	return f(dt)
}

func noopUpdate(_ float64) error {
	return nil
}
