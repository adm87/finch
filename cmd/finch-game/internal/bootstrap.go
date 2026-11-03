package internal

import (
	"github.com/adm87/finch/cmd/finch-game/internal/game/assets"
	"github.com/adm87/finch/cmd/finch-game/internal/game/draw"
	"github.com/adm87/finch/cmd/finch-game/internal/game/effects"
	"github.com/adm87/finch/cmd/finch-game/internal/game/scene"
	"github.com/adm87/finch/cmd/finch-game/internal/game/update"
	"github.com/adm87/finch/pkg/display"
	"github.com/adm87/finch/pkg/ecs"
	"github.com/adm87/finch/pkg/ecs/render"
	"github.com/adm87/finch/pkg/ecs/sprite"
	"github.com/adm87/finch/pkg/engine"
)

func Boot() error {
	renderPipeline := render.NewPipeline()
	world := ecs.NewWorld(renderPipeline)
	screen := display.NewScreen(400, 300)

	spriteKind := renderPipeline.Register(sprite.NewAdapter(assets.NewImages()))
	scene.LoadTest(world, spriteKind, screen.Visible(), 50)

	vignette, err := effects.NewVignette(1)
	if err != nil {
		return err
	}

	crt, err := effects.NewCRT(screen.Visible().Height)
	if err != nil {
		return err
	}

	drawer := draw.New(world, screen, vignette, crt)
	updater := update.New(world, vignette, crt)

	return engine.Run(
		engine.WithDrawer(drawer.Pipeline()),
		engine.WithUpdater(updater.UpdatePhase()),
		engine.WithLaterUpdater(updater.LaterPhase()),
		engine.WithDisplay(screen),
	)
}
