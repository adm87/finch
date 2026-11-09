package scene

import (
	"math"

	"github.com/adm87/finch/cmd/finch-game/internal/game/shuffle"
	"github.com/adm87/finch/pkg/ecs"
	"github.com/adm87/finch/pkg/ecs/render"
	"github.com/adm87/finch/pkg/ecs/sprite"
	"github.com/adm87/finch/pkg/ecs/transform"
	"github.com/adm87/finch/pkg/geom"
)

// LoadTest scatters count sprites across area.
func LoadTest(world *ecs.World, area geom.Rect, count int) {
	w := world.Donburi()
	for i := range count {
		e := w.Entry(w.Create(render.Component, sprite.Component, transform.Component, shuffle.Component))

		sprite.Component.Get(e).SetAnchor(0.5, 0.5)

		// The R2 sequence spreads sprites out evenly, and it's the same every run.
		n := float64(i + 1)
		_, fx := math.Modf(n * 0.7548776662466927)
		_, fy := math.Modf(n * 0.5698402909980532)
		x, y := area.X+fx*area.Width, area.Y+fy*area.Height
		transform.Component.Get(e).SetPosition(x, y)

		s := shuffle.Component.Get(e)
		s.SetHome(x, y)
		s.SetRadius(3 + 4*fy)
		s.SetSpeed(5)
		s.SetPhase(n)
	}
}
