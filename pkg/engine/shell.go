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
	s.clock.tick()

	dt := s.clock.dt.Seconds()
	fdt := s.clock.fixedDT.Seconds()

	if err := s.updater.Update(dt); err != nil {
		return err
	}
	for range s.clock.fixedSteps {
		if err := s.fixedUpdater.Update(fdt); err != nil {
			return err
		}
	}
	if err := s.laterUpdater.Update(dt); err != nil {
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
