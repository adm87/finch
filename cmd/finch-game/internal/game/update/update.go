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
	world *ecs.World
	crt   *effects.CRT
}

func New(world *ecs.World, crt *effects.CRT) *Updater {
	return &Updater{
		world: world,
		crt:   crt,
	}
}

// Main runs once per update, before the fixed steps.
func (u *Updater) Main() engine.Updater {
	return engine.UpdatePhase{
		quit(),
		u.toggleEffects(),
		u.shuffle(),
		u.crt,
		u.syncWorld(),
	}
}

func (u *Updater) shuffle() engine.UpdaterFunc {
	return func(dt float64) error {
		shuffle.Update(u.world.Donburi(), dt)
		return nil
	}
}

func (u *Updater) toggleEffects() engine.UpdaterFunc {
	return func(_ float64) error {
		if inpututil.IsKeyJustPressed(ebiten.Key1) {
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
