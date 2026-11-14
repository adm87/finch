package scene

import (
	"math"

	"github.com/adm87/finch/cmd/finch-game/internal/game/shuffle"
	"github.com/adm87/finch/pkg/ecs"
	"github.com/adm87/finch/pkg/ecs/bounds"
	"github.com/adm87/finch/pkg/ecs/rendering"
	"github.com/adm87/finch/pkg/ecs/sprite"
	"github.com/adm87/finch/pkg/ecs/transform"
	"github.com/adm87/finch/pkg/geom"
)

// Scatter spreads count sprites across area.
func Scatter(world *ecs.World, images sprite.Images, area geom.Rect, count int) {
	w := world.Donburi()
	for i := range count {
		e := w.Entry(w.Create(
			rendering.Component,
			sprite.Component,
			transform.Component,
			shuffle.Component,
			bounds.Component,
		))

		spr := sprite.Component.Get(e)
		spr.SetAnchor(0.5, 0.5)

		size := images.Get(spr.Handle()).Bounds().Size()
		width, height := float64(size.X), float64(size.Y)
		ax, ay := spr.Anchor()
		bounds.Component.Get(e).SetLocal(geom.Rect{X: -ax * width, Y: -ay * height, Width: width, Height: height})

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
