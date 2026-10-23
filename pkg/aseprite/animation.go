package aseprite

import (
	"math"

	"github.com/yohamta/donburi"

	"github.com/adm87/finch/pkg/images"
)

// AdvanceAnimations steps every entry that has an Animator and an Image.
func AdvanceAnimations(dt float64, libs Libraries, entries []*donburi.Entry) error {
	for _, entry := range entries {
		aComp, aOk := GetAnimator(entry)
		if !aOk || aComp.Library() == 0 || aComp.Animation() == "" {
			continue
		}

		iComp, iOk := images.GetImage(entry)
		if !iOk {
			continue
		}

		lib, err := libs.GetLibrary(aComp.Library())
		if err != nil {
			return err
		}

		if err := advanceAnimation(dt, lib, aComp, iComp); err != nil {
			return err
		}
	}
	return nil
}

func advanceAnimation(dt float64, library LibraryInfo, aComp *Animator, iComp *images.Image) error {
	tag := library.Atlas.Meta.FrameTags.ByName(aComp.Animation())
	if tag == nil {
		return nil
	}

	frames := library.Atlas.Frames.Slice(tag.From, tag.To+1)
	n := len(frames)
	if n == 0 {
		return nil
	}

	if aComp.playing != aComp.animation || aComp.frame < 0 || aComp.frame >= n {
		aComp.restart(tag, n)
	}
	if dt > 0 {
		stepFrames(dt*1000, frames, tag, aComp)
	}

	iComp.SetFrame(tag.From + aComp.frame)
	return nil
}

// stepFrames advances the animator by ms. A frame with no duration holds playback.
func stepFrames(ms float64, frames Frames, tag *FrameTag, aComp *Animator) {
	if aComp.finished {
		return
	}

	n := len(frames)
	reverse := isAnimationReverse(tag)
	pingPong := isAnimationPingPong(tag) && n > 1

	cycle := n
	if pingPong {
		cycle = 2 * (n - 1)
	}

	aComp.elapsed += ms

	// Skip whole cycles so a long update doesn't step through every frame.
	if aComp.step < cycle {
		if period := cycleDuration(frames, cycle, aComp.frame, reverse, pingPong, aComp.pingPongReversing); period > 0 {
			if aComp.loops == LoopForever {
				aComp.elapsed = math.Mod(aComp.elapsed, period)
			} else if skip := min(math.Floor(aComp.elapsed/period), float64(aComp.loops-aComp.cycles)); skip > 0 {
				aComp.elapsed -= skip * period
				aComp.cycles += int(skip)
			}
		}
	}

	for {
		// A held cycle that was given more loops wraps before playing on.
		if aComp.step == cycle {
			aComp.frame, aComp.pingPongReversing = nextFrame(aComp.frame, n, reverse, pingPong, aComp.pingPongReversing)
			aComp.step = 0
		}

		d := float64(frames[aComp.frame].Duration)
		if d <= 0 || aComp.elapsed < d {
			return
		}
		aComp.elapsed -= d

		if !pingPong && aComp.step == cycle-1 {
			aComp.step = cycle
			if aComp.completeCycle() {
				return
			}
			continue
		}

		aComp.frame, aComp.pingPongReversing = nextFrame(aComp.frame, n, reverse, pingPong, aComp.pingPongReversing)
		aComp.step++
		if aComp.step == cycle {
			aComp.step = 0
			if aComp.completeCycle() {
				return
			}
		}
	}
}

// completeCycle reports whether the last cycle just ended.
func (a *Animator) completeCycle() bool {
	a.cycles++
	if a.hasCyclesLeft() {
		return false
	}
	a.finished = true
	a.elapsed = 0
	return true
}

// cycleDuration returns 0 if a frame in the cycle has no duration.
func cycleDuration(frames Frames, cycle, idx int, reverse, pingPong, reversing bool) float64 {
	total := 0.0
	for range cycle {
		d := float64(frames[idx].Duration)
		if d <= 0 {
			return 0
		}
		total += d
		idx, reversing = nextFrame(idx, len(frames), reverse, pingPong, reversing)
	}
	return total
}

func isAnimationReverse(tag *FrameTag) bool {
	return tag.Direction == DirectionReverse || tag.Direction == DirectionPingPongReverse
}

func isAnimationPingPong(tag *FrameTag) bool {
	return tag.Direction == DirectionPingPong || tag.Direction == DirectionPingPongReverse
}

func nextFrame(idx, n int, reverse, pingPong, reversing bool) (int, bool) {
	if !pingPong {
		if reverse {
			return (idx - 1 + n) % n, reversing
		}
		return (idx + 1) % n, reversing
	}

	if idx <= 0 {
		reversing = false
	} else if idx >= n-1 {
		reversing = true
	}
	if reversing {
		return idx - 1, true
	}
	return idx + 1, false
}
