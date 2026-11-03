package engine

import (
	"slices"
	"testing"
)

func TestClockFixedSteps(t *testing.T) {
	tests := []struct {
		name                string
		tickRate, fixedRate int
		want                []int
	}{
		{"same rate", 60, 60, []int{1, 1, 1, 1}},
		{"half rate", 60, 30, []int{0, 1, 0, 1}},
		{"double rate", 60, 120, []int{2, 2, 2, 2}},
		{"uneven rate", 60, 50, []int{0, 1, 1, 1, 1, 1, 0, 1, 1, 1, 1, 1}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := newClock(tt.tickRate, tt.fixedRate)
			got := make([]int, len(tt.want))
			for i := range got {
				got[i] = c.advance()
			}
			if !slices.Equal(got, tt.want) {
				t.Errorf("fixed steps = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestClockStepsMatchRateOverASecond(t *testing.T) {
	c := newClock(60, 50)
	var steps int
	for range 60 {
		steps += c.advance()
	}
	if steps != 50 {
		t.Errorf("fixed steps in one second = %d, want 50", steps)
	}
}

func TestClockDeltas(t *testing.T) {
	c := newClock(60, 120)
	if c.dt != 1.0/60 || c.fixedDT != 1.0/120 {
		t.Errorf("dt = %v, fixedDT = %v, want %v and %v", c.dt, c.fixedDT, 1.0/60, 1.0/120)
	}
}

func TestClockRejectsZeroRates(t *testing.T) {
	c := newClock(0, 0)
	if got := c.advance(); got != 1 {
		t.Errorf("fixed steps = %d, want 1", got)
	}
}
