package update

import (
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"

	"github.com/adm87/finch/cmd/finch-game/internal/game/shuffle"
	"github.com/adm87/finch/pkg/ecs"
	"github.com/adm87/finch/pkg/engine"
)

type Updater struct {
	world *ecs.World
}

func New(world *ecs.World) *Updater {
	return &Updater{
		world: world,
	}
}

func (u *Updater) UpdatePhase() engine.Updater {
	return engine.UpdatePhase{
		quit(),
		u.shuffle(),
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
