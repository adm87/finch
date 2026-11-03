package engine

// clock counts updates and works out how many fixed steps are due. Ebitengine
// runs Update at a fixed TPS, so game time is just the tick count.
type clock struct {
	tick      int64
	tickRate  int64
	fixedRate int64
	dt        float64 // seconds per update
	fixedDT   float64 // seconds per fixed step
}

func newClock(tickRate, fixedRate int) clock {
	tickRate, fixedRate = max(tickRate, 1), max(fixedRate, 1)
	return clock{
		tickRate:  int64(tickRate),
		fixedRate: int64(fixedRate),
		dt:        1 / float64(tickRate),
		fixedDT:   1 / float64(fixedRate),
	}
}

// advance moves on one tick and returns how many fixed steps are now due.
func (c *clock) advance() int {
	prev := c.tick * c.fixedRate / c.tickRate
	c.tick++
	return int(c.tick*c.fixedRate/c.tickRate - prev)
}
