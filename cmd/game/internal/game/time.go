package game

import "time"

type Time struct {
	lastTick       time.Time
	deltaTime      time.Duration
	fixedDeltaTime time.Duration
	accumulator    time.Duration
	fixedSteps     int
}

const (
	AccumulatorMax = time.Second
	DeltaTimeMax   = time.Second
)

func NewTime(fps int) *Time {
	fps = max(fps, 1)
	return &Time{
		fixedDeltaTime: time.Second / time.Duration(fps),
	}
}

func (gt *Time) DeltaTime() float64 {
	return gt.deltaTime.Seconds()
}

func (gt *Time) FixedDeltaTime() float64 {
	return gt.fixedDeltaTime.Seconds()
}

func (gt *Time) FixedSteps() int {
	return gt.fixedSteps
}

func (gt *Time) Tick() {
	now := time.Now()

	if gt.lastTick.IsZero() {
		gt.lastTick = now
		return
	}

	gt.deltaTime = now.Sub(gt.lastTick)
	gt.lastTick = now

	if gt.deltaTime < 0 {
		gt.deltaTime = 0
	}

	if gt.deltaTime > DeltaTimeMax {
		gt.deltaTime = DeltaTimeMax
	}

	gt.accumulator += gt.deltaTime
	gt.fixedSteps = 0

	if gt.accumulator > AccumulatorMax {
		gt.accumulator = AccumulatorMax
	}

	for gt.accumulator >= gt.fixedDeltaTime {
		gt.accumulator -= gt.fixedDeltaTime
		gt.fixedSteps++
	}
}
