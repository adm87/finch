package internal

import (
	"github.com/adm87/finch/cmd/game/internal/camera"
	"github.com/adm87/finch/cmd/game/internal/game"
	"github.com/adm87/finch/cmd/game/internal/input"
	"github.com/adm87/finch/cmd/game/internal/player"
	"github.com/adm87/finch/pkg/ecs/systems"
	"github.com/adm87/finch/pkg/ecs/world"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/yohamta/donburi"
)

func initializeFrameSchedule(g *game.Game, w world.World, gu *game.GameUpdate) error {
	var entries []*donburi.Entry

	// Input
	gu.AddUpdate(func(dt float64) error {
		if inpututil.IsKeyJustPressed(ebiten.KeyEscape) {
			return ebiten.Termination
		}
		if c, cOk := camera.MainCamera.First(w.ECS()); cOk {
			input.UpdateCamera(dt, c)
		}
		if p, pOk := player.Player.First(w.ECS()); pOk {
			input.UpdatePlayer(dt, p, w)
		}
		return nil
	})

	// Collect entries for processing
	gu.AddLate(func(dt float64) error {
		rect, _ := g.View.Viewport()
		entries = w.Query(rect, entries[:0])
		return nil
	})

	// Animations
	gu.AddLate(func(dt float64) error {
		return systems.AdvanceAsepriteAnimations(dt, g.Aseprite, entries)
	})

	// Camera follow target
	gu.AddLate(func(dt float64) error {
		if c, cOk := camera.MainCamera.First(w.ECS()); cOk {
			if p, pOk := player.Player.First(w.ECS()); pOk {
				camera.FollowTarget(dt, c, p)
			}
		}
		return nil
	})

	return nil
}
