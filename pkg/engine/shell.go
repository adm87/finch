package engine

import (
	"github.com/hajimehoshi/ebiten/v2"

	"github.com/adm87/finch/pkg/application"
	"github.com/adm87/finch/pkg/logger"
)

type shell struct {
	app     *application.Application
	options *RunOptions
	clock   *clock
}

func newShell(app *application.Application, options *RunOptions, clock *clock) *shell {
	return &shell{app: app, options: options, clock: clock}
}

func (sh *shell) Update() error {
	sh.clock.tick()

	dt := sh.clock.delta.Seconds()
	if err := sh.runSystems(sh.options.Update, dt); err != nil {
		return err
	}
	for range sh.clock.fixedSteps {
		if err := sh.runSystems(sh.options.FixedUpdate, sh.clock.fixedDelta.Seconds()); err != nil {
			return err
		}
	}
	return sh.runSystems(sh.options.LateUpdate, dt)
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
	buf.Fill(sh.options.ClearColor)

	if err := sh.options.Drawer.Draw(buf, sh.app); err != nil {
		sh.options.Logger.Error("draw error", logger.ErrorField("error", err))
	}

	target.DrawImage(buf, screen.Options())

	if err := sh.options.Overlay.Draw(target, sh.app); err != nil {
		sh.options.Logger.Error("overlay error", logger.ErrorField("error", err))
	}
}

func (sh *shell) Layout(outsideWidth, outsideHeight int) (int, int) {
	return sh.app.Screen().Layout(outsideWidth, outsideHeight)
}
