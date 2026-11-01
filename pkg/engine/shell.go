package engine

import (
	"image/color"
	"math"

	"github.com/hajimehoshi/ebiten/v2"

	"github.com/adm87/finch/pkg/display"
	"github.com/adm87/finch/pkg/hexcolor"
	"github.com/adm87/finch/pkg/logger"
	"github.com/adm87/finch/pkg/world"
)

type shell struct {
	clearColor   color.Color
	world   *world.World
	screen  *display.Screen
	view    *display.View
	options *RunOptions
	clock   *clock
	drawCtx DrawContext
}

func newShell(w *world.World, screen *display.Screen, view *display.View, cfg Config, options *RunOptions, clock *clock) *shell {
	screen.SetResolutionScale(cfg.Window.ResolutionScale)
	screen.SetFilter(ebiten.Filter(cfg.Window.ScreenFilter))

	return &shell{
		world:      w,
		screen:     screen,
		view:       view,
		options:    options,
		clock:      clock,
		clearColor: hexcolor.ToRGBA(cfg.ClearColor),
	}
}

func (sh *shell) Update() error {
	sh.clock.tick()

	dt := sh.clock.delta.Seconds()
	if err := sh.runPhase(sh.options.Update, dt); err != nil {
		return err
	}

	dt = sh.clock.fixedDelta.Seconds()
	for range sh.clock.fixedSteps {
		if err := sh.runPhase(sh.options.FixedUpdate, dt); err != nil {
			return err
		}
	}

	dt = sh.clock.delta.Seconds()
	return sh.runPhase(sh.options.LateUpdate, dt)
}

// runPhase runs systems, then syncs the world for the next phase.
func (sh *shell) runPhase(systems []System, dt float64) error {
	if err := sh.runSystems(systems, dt); err != nil {
		return err
	}
	return sh.world.Sync()
}

func (sh *shell) runSystems(systems []System, dt float64) error {
	for _, system := range systems {
		if err := system(dt); err != nil {
			return err
		}
	}
	return nil
}

func (sh *shell) Draw(target *ebiten.Image) {
	buf := sh.screen.Buffer()
	buf.Fill(sh.clearColor)

	// No camera means nothing to draw in the world, but overlays still run.
	ctx := &sh.drawCtx
	*ctx = DrawContext{Entries: ctx.Entries[:0]}

	if sh.view.HasCamera() {
		ctx.View, ctx.Matrix = sh.view.Viewport()
		ctx.Entries = sh.world.Query(ctx.View, ctx.Entries)

		if err := sh.options.WorldDrawers.Draw(buf, ctx); err != nil {
			sh.options.Logger.Error("draw error", logger.ErrorField("error", err))
		}
	}

	target.DrawImage(buf, sh.screen.Options())

	ctx.Matrix.Concat(sh.screen.Options().GeoM)
	if err := sh.options.OverlayDrawers.Draw(target, ctx); err != nil {
		sh.options.Logger.Error("overlay error", logger.ErrorField("error", err))
	}
}

// Layout isn't called since LayoutF exists, but ebiten.Game needs it.
func (sh *shell) Layout(outsideWidth, outsideHeight int) (int, int) {
	w, h := sh.LayoutF(float64(outsideWidth), float64(outsideHeight))
	return int(math.Ceil(w)), int(math.Ceil(h))
}

func (sh *shell) LayoutF(outsideWidth, outsideHeight float64) (float64, float64) {
	return sh.screen.Layout(outsideWidth, outsideHeight)
}
