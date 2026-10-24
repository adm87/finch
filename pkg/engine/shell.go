package engine

import (
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"

	"github.com/adm87/finch/pkg/application"
	"github.com/adm87/finch/pkg/logger"
)

type shell struct {
	app        *application.Application
	updater    Updater
	drawer     Drawer
	overlay    Drawer
	clearColor color.Color
	logger     logger.Logger
}

func (sh *shell) Update() error {
	return sh.updater.Update(sh.app)
}

func (sh *shell) Draw(target *ebiten.Image) {
	screen := sh.app.Screen()
	buf := screen.Buffer()
	buf.Fill(sh.clearColor)

	if err := sh.drawer.Draw(buf, sh.app); err != nil {
		sh.logger.Error("draw error", logger.ErrorField("error", err))
	}

	target.DrawImage(buf, screen.Options())

	if err := sh.overlay.Draw(target, sh.app); err != nil {
		sh.logger.Error("overlay error", logger.ErrorField("error", err))
	}
}

func (sh *shell) Layout(outsideWidth, outsideHeight int) (int, int) {
	return sh.app.Screen().Layout(outsideWidth, outsideHeight)
}
