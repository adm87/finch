package game

import (
	"github.com/adm87/finch/cmd/game/internal/debug"
	"github.com/adm87/finch/pkg/engine"
	"github.com/adm87/finch/pkg/logger"
)

// Run starts the game loop with the game's systems and drawers.
func (g *Game) Run(cfg *engine.Config, l logger.Logger) error {
	return engine.Run(cfg, g.world, g.screen, g.view,
		engine.WithUpdate(
			debug.PollInput(g.world),
			g.handleInput,
		),
		engine.WithLateUpdate(
			g.followPlayer,
			g.collectVisible,
			g.animate,
		),
		engine.WithWorldDrawers(
			engine.DrawPipeline(g.pipeline),
		),
		engine.WithOverlayDrawers(
			debug.OverlayEntityBounds(g.world),
			debug.OverlayFPS(g.world),
			debug.OverlayPartitions(g.world),
			debug.OverlayZeroZero(g.world),
		),
		engine.WithLogger(l),
	)
}
