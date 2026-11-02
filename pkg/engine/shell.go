package engine

import "github.com/hajimehoshi/ebiten/v2"

// shell interfaces with Ebitengine to provide a platform for running Finch applications.
type shell struct {
	display      Display
	drawer       Drawer
	updater      Updater
	fixedUpdater Updater
	laterUpdater Updater
	clock        clock
}

func (s *shell) Update() error {
	s.clock.tick()

	dt := s.clock.dt.Seconds()
	fdt := s.clock.fixedDT.Seconds()

	if err := update(s.updater, dt); err != nil {
		return err
	}
	for range s.clock.fixedSteps {
		if err := update(s.fixedUpdater, fdt); err != nil {
			return err
		}
	}
	if err := update(s.laterUpdater, dt); err != nil {
		return err
	}

	return nil
}

func (s *shell) Draw(screen *ebiten.Image) {
	s.drawer.Draw(screen)
}

func (s *shell) Layout(outsideWidth, outsideHeight int) (int, int) {
	return s.display.Layout(outsideWidth, outsideHeight)
}

func update(updater Updater, dt float64) error {
	if updater == nil {
		return nil
	}
	return updater.Update(dt)
}
