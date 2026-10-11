package engine

import (
	"github.com/adm87/finch/pkg/logger"
	"github.com/hajimehoshi/ebiten/v2"
)

type shell struct {
	update GameUpdate
	draw   GameDraw
	layout GameLayout
	logger logger.Logger
}

func (sh shell) Update() error {
	return sh.update.Update()
}

func (sh shell) Draw(screen *ebiten.Image) {
	if err := sh.draw.Draw(screen); err != nil {
		sh.logger.Error("draw error", logger.ErrorField("error", err))
	}
}

func (sh shell) Layout(outsideWidth, outsideHeight int) (int, int) {
	return sh.layout.Layout(outsideWidth, outsideHeight)
}
