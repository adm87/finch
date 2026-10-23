package camera

import (
	"github.com/yohamta/donburi"

	"github.com/adm87/finch/pkg/transform"
)

func FollowTarget(dt float64, c *donburi.Entry, target *donburi.Entry) {
	t := transform.TransformComponent.Get(c)
	tx, ty := transform.TransformComponent.Get(target).Position()
	x, y := t.Position()
	x += (tx - x) * dt
	y += (ty - y) * dt
	t.SetPosition(x, y)
}
