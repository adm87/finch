package draw

import (
	"strconv"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"

	"github.com/adm87/finch/pkg/display"
	"github.com/adm87/finch/pkg/ecs"
	"github.com/adm87/finch/pkg/ecs/render"
	"github.com/adm87/finch/pkg/engine"
	"github.com/adm87/finch/pkg/post"
)

type Drawer struct {
	world    *ecs.World
	screen   *display.Screen
	vignette *ebiten.Shader
}

func New(world *ecs.World, screen *display.Screen, vignette *ebiten.Shader) *Drawer {
	return &Drawer{
		world:    world,
		screen:   screen,
		vignette: vignette,
	}
}

func (d *Drawer) Pipeline() engine.Drawer {
	return engine.DrawPipeline(
		d.drawWorld(),
		post.Shader(d.vignette, map[string]any{"Strength": float32(1)}),
		d.screen.Present(),
		// TODO: Add UI pass
		d.debugFPS(),
	)
}

func (d *Drawer) drawWorld() engine.DrawerFunc {
	return func(target *ebiten.Image) {
		// TODO: Get the view from a camera
		d.world.Render(target, render.View{Bounds: d.screen.Visible()})
	}
}

func (d *Drawer) debugFPS() engine.DrawPass {
	return func(inner engine.Drawer) engine.Drawer {
		return engine.DrawerFunc(func(target *ebiten.Image) {
			inner.Draw(target)
			ebitenutil.DebugPrint(target, "FPS: "+strconv.Itoa(int(ebiten.ActualFPS())))
		})
	}
}
