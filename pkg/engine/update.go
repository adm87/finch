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

type UpdatePhase []Updater

func (u UpdatePhase) Update(dt float64) error {
	for _, updater := range u {
		if err := updater.Update(dt); err != nil {
			return err
		}
	}
	return nil
}

func noopUpdate(_ float64) error {
	return nil
}
