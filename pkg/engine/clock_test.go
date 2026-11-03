package engine

import (
	"testing"
	"time"
)

const testFixedDT = time.Second / 60

func TestClockFirstTickHasNoDelta(t *testing.T) {
	c := clock{fixedDT: testFixedDT}
	c.tick()

	if c.dt != 0 || c.fixedSteps != 0 {
		t.Errorf("dt = %v, fixedSteps = %d, want 0 and 0", c.dt, c.fixedSteps)
	}
}

func TestClockAccumulatesFixedSteps(t *testing.T) {
	c := clock{fixedDT: testFixedDT, last: time.Now().Add(-40 * time.Millisecond)}
	c.tick()

	if c.fixedSteps != 2 {
		t.Errorf("fixedSteps = %d, want 2", c.fixedSteps)
	}
	// The leftover carries into the next tick.
	if c.accumulated <= 0 || c.accumulated >= testFixedDT {
		t.Errorf("accumulated = %v, want between 0 and %v", c.accumulated, testFixedDT)
	}
}

func TestClockClampsLongFrames(t *testing.T) {
	c := clock{fixedDT: testFixedDT, last: time.Now().Add(-5 * time.Second)}
	c.tick()

	if c.dt != maxDelta {
		t.Errorf("dt = %v, want %v", c.dt, maxDelta)
	}
	if want := int(maxAccumulated / testFixedDT); c.fixedSteps != want {
		t.Errorf("fixedSteps = %d, want %d", c.fixedSteps, want)
	}
}

func TestClockIgnoresTimeGoingBackwards(t *testing.T) {
	c := clock{fixedDT: testFixedDT, last: time.Now().Add(time.Hour)}
	c.tick()

	if c.dt != 0 || c.fixedSteps != 0 {
		t.Errorf("dt = %v, fixedSteps = %d, want 0 and 0", c.dt, c.fixedSteps)
	}
}
