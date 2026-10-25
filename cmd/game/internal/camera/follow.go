package camera

import (
	"math"

	"github.com/yohamta/donburi"

	"github.com/adm87/finch/pkg/transform"
)

// followRate is how quickly the camera closes the gap to its target, per second.
const followRate = 1.0

func Follow(dt float64, c *donburi.Entry, target *donburi.Entry) {
	t, ok := transform.GetTransform(c)
	if !ok {
		return
	}
	tt, ok := transform.GetTransform(target)
	if !ok {
		return
	}

	tx, ty := tt.Position()
	x, y := t.Position()

	step := 1 - math.Exp(-followRate*dt)
	t.SetPosition(x+(tx-x)*step, y+(ty-y)*step)
}
