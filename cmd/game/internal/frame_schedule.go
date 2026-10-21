package internal

import (
	"github.com/adm87/finch/cmd/game/internal/camera"
	"github.com/adm87/finch/cmd/game/internal/game"
	"github.com/adm87/finch/cmd/game/internal/input"
	"github.com/adm87/finch/pkg/ecs/world"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
)

func initializeFrameSchedule(g *game.Game, w world.World, gu *game.GameUpdate) error {
	gu.AddUpdate(func(dt float64) error {
		if inpututil.IsKeyJustPressed(ebiten.KeyEscape) {
			return ebiten.Termination
		}
		return nil
	})

	gu.AddUpdate(func(dt float64) error {
		if c, cOk := camera.MainCamera.First(w.ECS()); cOk {
			input.UpdateCamera(dt, c)
		}
		return nil
	})

	return nil
}
