package engine

import "time"

const (
	maxDelta       = time.Second
	maxAccumulated = time.Second
)

// clock measures the time between updates and how many fixed steps are due.
type clock struct {
	last        time.Time
	dt          time.Duration
	fixedDT     time.Duration
	accumulated time.Duration
	fixedSteps  int
}

func (c *clock) tick() {
	now := time.Now()
	c.fixedSteps = 0

	if c.last.IsZero() {
		c.last = now
		c.dt = 0
		return
	}

	c.dt = min(max(now.Sub(c.last), 0), maxDelta)
	c.last = now

	c.accumulated = min(c.accumulated+c.dt, maxAccumulated)
	for c.accumulated >= c.fixedDT {
		c.accumulated -= c.fixedDT
		c.fixedSteps++
	}
}
