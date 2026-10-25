package engine

import (
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"

	"github.com/adm87/finch/pkg/application"
	"github.com/adm87/finch/pkg/hexcolor"
	"github.com/adm87/finch/pkg/logger"
)

type shell struct {
	clearColor   color.Color
	app          *application.Application
	options      *RunOptions
	clock        *clock
	worldDrawers Drawers
	drawCtx      DrawContext
}

func newShell(app *application.Application, cfg Config, options *RunOptions, clock *clock) *shell {
	worldDrawers := options.WorldDrawers
	if len(worldDrawers) == 0 {
		worldDrawers = Drawers{DrawerFunc(drawPipeline(app))}
	}
	return &shell{
		app:          app,
		options:      options,
		clock:        clock,
		worldDrawers: worldDrawers,
		clearColor:   hexcolor.ToRGBA(cfg.ClearColor),
	}
}

func drawPipeline(app *application.Application) func(*ebiten.Image, *DrawContext) error {
	return func(target *ebiten.Image, ctx *DrawContext) error {
		return app.RenderPipeline().Draw(target, ctx.Entries, ctx.View, ctx.Matrix)
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
	return sh.app.World().Sync()
}

func (sh *shell) runSystems(systems []System, dt float64) error {
	for _, system := range systems {
		if err := system(sh.app, dt); err != nil {
			return err
		}
	}
	return nil
}

func (sh *shell) Draw(target *ebiten.Image) {
	screen := sh.app.Screen()

	buf := screen.Buffer()
	buf.Fill(sh.clearColor)

	// No camera means nothing to draw in the world, but overlays still run.
	ctx := &sh.drawCtx
	*ctx = DrawContext{Entries: ctx.Entries[:0]}

	if sh.app.View().HasCamera() {
		ctx.View, ctx.Matrix = sh.app.View().Viewport()
		ctx.Entries = sh.app.World().Query(ctx.View, ctx.Entries)

		if err := sh.worldDrawers.Draw(buf, ctx); err != nil {
			sh.options.Logger.Error("draw error", logger.ErrorField("error", err))
		}
	}

	target.DrawImage(buf, screen.Options())

	ctx.Matrix.Concat(screen.Options().GeoM)
	if err := sh.options.OverlayDrawers.Draw(target, ctx); err != nil {
		sh.options.Logger.Error("overlay error", logger.ErrorField("error", err))
	}
}

func (sh *shell) Layout(outsideWidth, outsideHeight int) (int, int) {
	return sh.app.Screen().Layout(outsideWidth, outsideHeight)
}
