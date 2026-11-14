package update

import (
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"

	"github.com/adm87/finch/cmd/finch-game/internal/game/camera"
	"github.com/adm87/finch/cmd/finch-game/internal/game/effects"
	"github.com/adm87/finch/cmd/finch-game/internal/game/shuffle"
	"github.com/adm87/finch/pkg/display"
	"github.com/adm87/finch/pkg/ecs"
	"github.com/adm87/finch/pkg/engine"
)

type Updater struct {
	world  *ecs.World
	crt    *effects.CRT
	screen *display.Screen
}

func New(world *ecs.World, crt *effects.CRT, screen *display.Screen) *Updater {
	return &Updater{
		world:  world,
		crt:    crt,
		screen: screen,
	}
}

// Main runs once per update, before the fixed steps.
func (u *Updater) Main() engine.Updater {
	return engine.UpdatePhase{
		quit(),
		u.panCamera(),
		u.easeCamera(),
		u.fitCamera(),
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

func (u *Updater) easeCamera() engine.UpdaterFunc {
	return func(dt float64) error {
		camera.Ease(u.world.Donburi(), dt)
		return nil
	}
}

func (u *Updater) fitCamera() engine.UpdaterFunc {
	return func(_ float64) error {
		camera.Fit(u.world.Donburi(), u.screen)
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

func (u *Updater) panCamera() engine.UpdaterFunc {
	const speed = 600
	return func(dt float64) error {
		var dx, dy float64
		if ebiten.IsKeyPressed(ebiten.KeyW) {
			dy -= speed * dt
		}
		if ebiten.IsKeyPressed(ebiten.KeyS) {
			dy += speed * dt
		}
		if ebiten.IsKeyPressed(ebiten.KeyA) {
			dx -= speed * dt
		}
		if ebiten.IsKeyPressed(ebiten.KeyD) {
			dx += speed * dt
		}
		camera.Move(u.world.Donburi(), dx, dy)
		return nil
	}
}
