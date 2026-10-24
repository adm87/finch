package camera

import (
	"math"

	"github.com/yohamta/donburi"

	"github.com/adm87/finch/pkg/transform"
)

// followRate is how quickly the camera closes the gap to its target, per second.
const followRate = 1.0

func FollowTarget(dt float64, c *donburi.Entry, target *donburi.Entry) {
	t := transform.TransformComponent.Get(c)
	tx, ty := transform.TransformComponent.Get(target).Position()
	x, y := t.Position()

	step := 1 - math.Exp(-followRate*dt)
	t.SetPosition(x+(tx-x)*step, y+(ty-y)*step)
}
