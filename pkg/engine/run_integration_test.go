//go:build integration

package engine_test

import (
	"testing"

	"github.com/hajimehoshi/ebiten/v2"

	"github.com/adm87/finch/pkg/engine"
)

func TestRunDrawsAndStops(t *testing.T) {
	var updates, draws int

	err := engine.Run(
		engine.WithUpdater(engine.UpdaterFunc(func(float64) error {
			updates++
			if draws > 0 || updates > 600 {
				return ebiten.Termination
			}
			return nil
		})),
		engine.WithDrawer(engine.DrawerFunc(func(*ebiten.Image) { draws++ })),
	)
	if err != nil {
		t.Fatalf("Run() = %v", err)
	}
	if draws == 0 {
		t.Errorf("Run never drew a frame in %d updates", updates)
	}
}