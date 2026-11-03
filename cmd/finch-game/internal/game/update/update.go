package update

import (
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"

	"github.com/adm87/finch/cmd/finch-game/internal/game/effects"
	"github.com/adm87/finch/cmd/finch-game/internal/game/shuffle"
	"github.com/adm87/finch/pkg/ecs"
	"github.com/adm87/finch/pkg/engine"
)

type Updater struct {
	world    *ecs.World
	vignette *effects.Vignette
	crt      *effects.CRT
}

func New(world *ecs.World, vignette *effects.Vignette, crt *effects.CRT) *Updater {
	return &Updater{
		world:    world,
		vignette: vignette,
		crt:      crt,
	}
}

func (u *Updater) UpdatePhase() engine.Updater {
	return engine.UpdatePhase{
		quit(),
		u.toggleEffects(),
		u.shuffle(),
		u.crt,
	}
}

func (u *Updater) LaterPhase() engine.Updater {
	return engine.UpdatePhase{
		u.syncWorld(),
	}
}

func (u *Updater) shuffle() engine.UpdaterFunc {
	return func(dt float64) error {
		shuffle.Update(u.world.Donburi(), dt)
		return nil
	}
}

// 1 toggles the vignette, 2 toggles the CRT.
func (u *Updater) toggleEffects() engine.UpdaterFunc {
	return func(_ float64) error {
		if inpututil.IsKeyJustPressed(ebiten.Key1) {
			u.vignette.Toggle()
		}
		if inpututil.IsKeyJustPressed(ebiten.Key2) {
			u.crt.Toggle()
		}
		return nil
	}
}

func (u *Updater) syncWorld() engine.UpdaterFunc {
	return func(_ float64) error {
		u.world.Sync()
		return nil
	}
}

func quit() engine.UpdaterFunc {
	return func(_ float64) error {
		if inpututil.IsKeyJustPressed(ebiten.KeyEscape) {
			return ebiten.Termination
		}
		return nil
	}
}
