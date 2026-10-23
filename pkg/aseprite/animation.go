package aseprite

import (
	"github.com/yohamta/donburi"

	"github.com/adm87/finch/pkg/images"
)

// AdvanceAnimations steps every entry that has an Animator and an Image.
func AdvanceAnimations(dt float64, libs Libraries, entries []*donburi.Entry) error {
	for _, entry := range entries {
		anim, ok := GetAnimator(entry)
		if !ok || anim.Library() == 0 || anim.Animation() == "" {
			continue
		}

		img, ok := images.GetImage(entry)
		if !ok {
			continue
		}

		lib, err := libs.GetLibrary(anim.Library())
		if err != nil {
			return err
		}

		advanceAnimation(dt, lib, anim, img)
	}
	return nil
}

func advanceAnimation(dt float64, library LibraryInfo, anim *Animator, img *images.Image) {
	tag := library.Atlas.Meta.FrameTags.ByName(anim.Animation())
	if tag == nil {
		return
	}

	frames := library.Atlas.Frames.Slice(tag.From, tag.To+1)

	if anim.playing != anim.animation || anim.position > cycleLength(len(frames), tag.Direction) {
		restartAnimation(anim)
	}
	if dt > 0 {
		advanceFrames(anim, dt*1000, frames, tag.Direction)
	}

	anim.frame = cycleFrame(anim.position, len(frames), tag.Direction)
	img.SetFrame(tag.From + anim.frame)
}

func restartAnimation(a *Animator) {
	a.playing = a.animation
	a.position = 0
	a.elapsed = 0
	a.cycles = 0
	a.finished = false
}

func hasCyclesLeft(a *Animator) bool {
	return a.loops == LoopForever || a.cycles <= a.loops
}

// advanceFrames moves playback forward by ms. A frame with no duration holds playback.
func advanceFrames(a *Animator, ms float64, frames Frames, dir Direction) {
	if a.finished && !hasCyclesLeft(a) {
		return
	}

	a.finished = false
	a.elapsed += ms

	length := cycleLength(len(frames), dir)
	if a.position == length {
		a.position = 0
	}

	for !a.finished {
		d := float64(frames[cycleFrame(a.position, len(frames), dir)].Duration)
		if d <= 0 || a.elapsed < d {
			return
		}

		a.elapsed -= d
		a.position++

		if a.position == length {
			endCycle(a)
		}
	}
}

// endCycle starts the next cycle, or holds at the end if this was the last one.
func endCycle(a *Animator) {
	a.cycles++
	if hasCyclesLeft(a) {
		a.position = 0
		return
	}
	a.finished = true
	a.elapsed = 0
}

// cycleLength is the number of frames shown in one cycle. Ping-pong doesn't
// repeat its end frames.
func cycleLength(n int, dir Direction) int {
	if n > 1 && (dir == DirectionPingPong || dir == DirectionPingPongReverse) {
		return 2 * (n - 1)
	}
	return n
}

// cycleFrame maps a position in the cycle to a frame index. A finished
// animation holds at position cycleLength, which shows the last frame, or the
// starting frame for ping-pong.
func cycleFrame(position, n int, dir Direction) int {
	last := n - 1
	if last == 0 {
		return 0
	}
	switch dir {
	case DirectionReverse:
		return max(last-position, 0)
	case DirectionPingPong:
		if position <= last {
			return position
		}
		return 2*last - position
	case DirectionPingPongReverse:
		if position <= last {
			return last - position
		}
		return position - last
	default:
		return min(position, last)
	}
}
