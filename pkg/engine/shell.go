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
}

func newShell(app *application.Application, options *RunOptions, clock *clock) *shell {
	worldDrawers := options.WorldDrawers
	if len(worldDrawers) == 0 {
		worldDrawers = Drawers{DrawerFunc(drawPipeline)}
	}
	return &shell{
		app:          app,
		options:      options,
		clock:        clock,
		worldDrawers: worldDrawers,
		clearColor:   hexcolor.ToRGBA(app.Config().Screen.ClearColor),
	}
}

func drawPipeline(target *ebiten.Image, app *application.Application) error {
	return app.RenderPipeline().Draw(target)
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

// runPhase runs systems, then syncs the world so the next phase sees their changes.
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

	if err := sh.worldDrawers.Draw(buf, sh.app); err != nil {
		sh.options.Logger.Error("draw error", logger.ErrorField("error", err))
	}

	target.DrawImage(buf, screen.Options())

	if err := sh.options.OverlayDrawers.Draw(target, sh.app); err != nil {
		sh.options.Logger.Error("overlay error", logger.ErrorField("error", err))
	}
}

func (sh *shell) Layout(outsideWidth, outsideHeight int) (int, int) {
	return sh.app.Screen().Layout(outsideWidth, outsideHeight)
}
