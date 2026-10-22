package camera

import (
	"github.com/adm87/finch/pkg/ecs/components"
	"github.com/yohamta/donburi"
)

func FollowTarget(dt float64, c *donburi.Entry, target *donburi.Entry) {
	t := components.TransformComponent.Get(c)
	tx, ty := components.TransformComponent.Get(target).Position()
	x, y := t.Position()
	x += (tx - x) * dt
	y += (ty - y) * dt
	t.SetPosition(x, y)
}
