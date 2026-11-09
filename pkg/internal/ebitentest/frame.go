// Package ebitentest runs test code inside a real Ebitengine frame.
package ebitentest

import (
	"testing"

	"github.com/hajimehoshi/ebiten/v2"

	"github.com/adm87/finch/pkg/engine"
)

// InFrame runs f inside the game's first drawn frame. Pixels can only be read
// back once the game is running, and Run can only be called once per process.
func InFrame(t testing.TB, f func()) {
	t.Helper()
	var updates int
	var drawn bool
	err := engine.Run(
		engine.WithUpdater(engine.UpdaterFunc(func(float64) error {
			updates++
			if drawn || updates > 600 {
				return ebiten.Termination
			}
			return nil
		})),
		engine.WithDrawer(engine.DrawFunc(func(*ebiten.Image) {
			if !drawn {
				f()
				drawn = true
			}
		})),
	)
	if err != nil {
		t.Fatalf("Run() = %v", err)
	}
	if !drawn {
		t.Fatalf("Run never drew a frame in %d updates", updates)
	}
}
