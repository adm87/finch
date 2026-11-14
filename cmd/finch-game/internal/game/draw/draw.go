package draw

import (
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"

	"github.com/adm87/finch/cmd/finch-game/internal/devinfo"
	"github.com/adm87/finch/cmd/finch-game/internal/game/camera"
	"github.com/adm87/finch/pkg/display"
	"github.com/adm87/finch/pkg/ecs"
	"github.com/adm87/finch/pkg/engine"
)

type Drawer struct {
	world  *ecs.World
	screen *display.Screen
}

func New(world *ecs.World, screen *display.Screen) *Drawer {
	return &Drawer{
		world:  world,
		screen: screen,
	}
}

func (d *Drawer) Pipeline() engine.Drawer {
	return engine.DrawPipeline(
		d.drawWorld(),
		d.drawUI(),
		d.drawToScreen(),
		devinfo.DrawWorldZero(d.world.Donburi(), d.screen),
		devinfo.DrawInfo(),
	)
}

func (d *Drawer) drawWorld() engine.DrawFunc {
	return func(target *ebiten.Image) {
		d.world.Render(target, camera.View(d.world.Donburi(), d.screen.Visible()))
	}
}

func (d *Drawer) drawUI() engine.DrawPass {
	const margin = 10
	return func(inner engine.Drawer) engine.Drawer {
		return engine.DrawFunc(func(target *ebiten.Image) {
			inner.Draw(target)

			visible := d.screen.Visible()

			minX, minY := visible.Min()
			maxX, maxY := visible.Max()

			vector.FillRect(target, float32(minX), float32(minY), margin, margin, color.RGBA{R: 255, A: 255}, false)
			vector.FillRect(target, float32(maxX-margin), float32(minY), margin, margin, color.RGBA{G: 255, A: 255}, false)
			vector.FillRect(target, float32(minX), float32(maxY-margin), margin, margin, color.RGBA{B: 255, A: 255}, false)
			vector.FillRect(target, float32(maxX-margin), float32(maxY-margin), margin, margin, color.RGBA{R: 255, G: 255, A: 255}, false)
		})
	}
}

func (d *Drawer) drawToScreen() engine.DrawPass {
	return d.screen.Present()
}
