package aseprite

import (
	"github.com/yohamta/donburi"
)

// LoopForever is the loop count for an animation that never finishes.
const LoopForever = -1

// Animator plays an animation from an Aseprite library on an entity's image.
type Animator struct {
	animation string
	playing   string
	library   Handle
	loops     int
	position  int
	frame     int
	elapsed   float64
	cycles    int
	finished  bool
}

var AnimatorComponent = donburi.NewComponentType[Animator](Animator{
	loops: LoopForever,
})

func GetAnimator(entry *donburi.Entry) (*Animator, bool) {
	if !entry.HasComponent(AnimatorComponent) {
		return nil, false
	}

	return AnimatorComponent.Get(entry), true
}

func (a *Animator) Animation() string {
	return a.animation
}

// SetAnimation does not restart an animation that is already playing.
func (a *Animator) SetAnimation(animation string) {
	a.animation = animation
}

func (a *Animator) Library() Handle {
	return a.library
}

func (a *Animator) SetLibrary(library Handle) {
	a.library = library
	a.playing = ""
}

func (a *Animator) Loops() int {
	return a.loops
}

// SetLoops sets how many times the animation repeats after its first cycle.
// LoopForever never stops and 0 plays once. Ping-pong cycles go out and back.
func (a *Animator) SetLoops(loops int) {
	a.loops = max(loops, LoopForever)
}

// Restart plays the animation again from the start.
func (a *Animator) Restart() {
	a.playing = ""
}

func (a *Animator) Finished() bool {
	return a.finished
}

// Frame returns the frame index within the animation.
func (a *Animator) Frame() int {
	return a.frame
}

// Elapsed returns milliseconds spent on the current frame.
func (a *Animator) Elapsed() float64 {
	return a.elapsed
}
