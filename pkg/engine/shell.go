package engine

import "github.com/hajimehoshi/ebiten/v2"

// shell is the ebiten.Game that drives Finch.
type shell struct {
	display      Display
	drawer       Drawer
	updater      Updater
	fixedUpdater Updater
	laterUpdater Updater
	clock        clock
}

func (s *shell) Update() error {
	steps := s.clock.advance()

	if err := s.updater.Update(s.clock.dt); err != nil {
		return err
	}
	for range steps {
		if err := s.fixedUpdater.Update(s.clock.fixedDT); err != nil {
			return err
		}
	}
	if err := s.laterUpdater.Update(s.clock.dt); err != nil {
		return err
	}

	return nil
}

func (s *shell) Draw(screen *ebiten.Image) {
	s.drawer.Draw(screen)
}

func (s *shell) Layout(outsideWidth, outsideHeight int) (int, int) {
	w, h := s.LayoutF(float64(outsideWidth), float64(outsideHeight))
	return int(w), int(h)
}

func (s *shell) LayoutF(outsideWidth, outsideHeight float64) (float64, float64) {
	return s.display.Layout(outsideWidth, outsideHeight)
}
