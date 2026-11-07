package draw

import (
	"github.com/hajimehoshi/ebiten/v2"

	"github.com/adm87/finch/cmd/finch-game/internal/debug"
	"github.com/adm87/finch/cmd/finch-game/internal/game/effects"
	"github.com/adm87/finch/pkg/display"
	"github.com/adm87/finch/pkg/ecs"
	"github.com/adm87/finch/pkg/ecs/render"
	"github.com/adm87/finch/pkg/engine"
)

type Drawer struct {
	world    *ecs.World
	screen   *display.Screen
	vignette *effects.Vignette
	crt      *effects.CRT
}

func New(world *ecs.World, screen *display.Screen, vignette *effects.Vignette, crt *effects.CRT) *Drawer {
	return &Drawer{
		world:    world,
		screen:   screen,
		vignette: vignette,
		crt:      crt,
	}
}

func (d *Drawer) Pipeline() engine.Drawer {
	return engine.DrawPipeline(
		d.drawWorld(),
		d.drawVignette(),
		d.drawToScreen(),
		d.drawCRT(),
		d.drawUI(),
		debug.DrawInfo(),
	)
}

func (d *Drawer) drawWorld() engine.DrawerFunc {
	return func(target *ebiten.Image) {
		// TODO: Get the view from a camera
		d.world.Render(target, render.View{Bounds: d.screen.Visible()})
	}
}

func (d *Drawer) drawUI() engine.DrawPass {
	return func(inner engine.Drawer) engine.Drawer {
		return engine.DrawerFunc(func(target *ebiten.Image) {
			inner.Draw(target)
			// TODO: Draw UI elements here
		})
	}
}

func (d *Drawer) drawVignette() engine.DrawPass {
	return d.vignette.Pass()
}

func (d *Drawer) drawCRT() engine.DrawPass {
	return d.crt.Pass()
}

func (d *Drawer) drawToScreen() engine.DrawPass {
	return d.screen.Present()
}
