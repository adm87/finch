package engine

import "time"

const (
	maxDelta       = time.Second
	maxAccumulated = time.Second
)

// clock measures the time between updates and how many fixed steps are due.
type clock struct {
	now         func() time.Time
	last        time.Time
	delta       time.Duration
	fixedDelta  time.Duration
	accumulated time.Duration
	fixedSteps  int
}

func newClock(tickRate int, now func() time.Time) *clock {
	return &clock{
		now:        now,
		fixedDelta: time.Second / time.Duration(max(tickRate, 1)),
	}
}

func (c *clock) tick() {
	now := c.now()
	c.fixedSteps = 0

	if c.last.IsZero() {
		c.last = now
		c.delta = 0
		return
	}

	c.delta = min(max(now.Sub(c.last), 0), maxDelta)
	c.last = now

	c.accumulated = min(c.accumulated+c.delta, maxAccumulated)
	for c.accumulated >= c.fixedDelta {
		c.accumulated -= c.fixedDelta
		c.fixedSteps++
	}
}
