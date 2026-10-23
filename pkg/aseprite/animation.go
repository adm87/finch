package aseprite

import (
	"math"

	"github.com/yohamta/donburi"

	"github.com/adm87/finch/pkg/images"
)

// AdvanceAnimations steps the animation of every entry that has both an Animator and an Image.
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
	if dt <= 0 {
		return nil
	}

	tag := library.Atlas.Meta.FrameTags.ByName(aComp.Animation())
	if tag == nil {
		return nil
	}

	frames := library.Atlas.Frames.Slice(tag.From, tag.To+1)
	n := len(frames)
	if n == 0 {
		return nil
	}

	reverse := isAnimationReverse(tag)
	pingPong := isAnimationPingPong(tag)

	idx := iComp.Frame()
	if idx < 0 || idx >= n {
		idx = 0
		if reverse {
			idx = n - 1
		}
		iComp.SetFrame(idx)
		aComp.SetPingPongReversing(tag.Direction == DirectionPingPongReverse)
	}

	elapsed := aComp.Elapsed() + dt*1000
	duration := float64(frames[idx].Duration)

	if duration <= 0 || elapsed < duration {
		aComp.SetElapsed(elapsed)
		return nil
	}

	aComp.SetElapsed(math.Mod(elapsed, duration))

	if n == 1 {
		return nil
	}

	cycle := n
	if pingPong {
		cycle = 2 * (n - 1)
	}
	steps := int(math.Mod(elapsed/duration, float64(cycle)))

	reversing := aComp.PingPongReversing()
	for range steps {
		idx, reversing = nextFrame(idx, n, reverse, pingPong, reversing)
	}

	iComp.SetFrame(idx)
	aComp.SetPingPongReversing(reversing)
	return nil
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
