package main

import (
	"log"

	"github.com/hajimehoshi/ebiten/v2"

	"github.com/adm87/finch/cmd/finch/internal/game"
)

const (
	screenWidth, screenHeight = 320, 180
	windowSizeMultiplier      = 4
	windowTitle               = "Finch"
)

func main() {
	g, err := game.NewGame(screenWidth, screenHeight)
	if err != nil {
		log.Fatal(err)
	}

	ebiten.SetWindowTitle(windowTitle)
	ebiten.SetWindowSize(screenWidth*windowSizeMultiplier, screenHeight*windowSizeMultiplier)
	ebiten.SetWindowResizingMode(ebiten.WindowResizingModeEnabled)

	if err := ebiten.RunGame(g); err != nil {
		log.Fatal(err)
	}
}
