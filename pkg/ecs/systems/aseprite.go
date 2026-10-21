package systems

import (
	"math"

	"github.com/adm87/finch/pkg/aseprite"
	"github.com/adm87/finch/pkg/ecs/components"
	"github.com/yohamta/donburi"
)

func AdvanceAsepriteAnimations(dt float64, ase aseprite.Aseprite, entries []*donburi.Entry) error {
	for _, entry := range entries {
		aComp, aOk := components.GetAseprite(entry)
		if !aOk || aComp.Library() == 0 || aComp.Animation() == "" {
			continue
		}

		iComp, iOk := components.GetImage(entry)
		if !iOk {
			continue
		}

		lib, err := ase.GetLibrary(aseprite.LibraryID(aComp.Library()))
		if err != nil {
			return err
		}

		if err := advanceAsepriteAnimation(dt, lib, aComp, iComp); err != nil {
			return err
		}
	}
	return nil
}
func advanceAsepriteAnimation(dt float64, library aseprite.LibraryInfo, aComp *components.Aseprite, iComp *components.Image) error {
	if dt <= 0 || math.IsNaN(dt) || math.IsInf(dt, 0) {
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
		aComp.SetPingPongReversing(tag.Direction == aseprite.DirectionPingPongReverse)
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

func isAnimationReverse(tag *aseprite.FrameTag) bool {
	return tag.Direction == aseprite.DirectionReverse || tag.Direction == aseprite.DirectionPingPongReverse
}

func isAnimationPingPong(tag *aseprite.FrameTag) bool {
	return tag.Direction == aseprite.DirectionPingPong || tag.Direction == aseprite.DirectionPingPongReverse
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
