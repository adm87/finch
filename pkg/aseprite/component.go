package aseprite

import (
	"github.com/yohamta/donburi"
)

// Animator plays an animation from an Aseprite library on an entity's image.
type Animator struct {
	animation         string
	library           LibraryID
	elapsed           float64
	pingPongReversing bool
}

var AnimatorComponent = donburi.NewComponentType[Animator]()

func GetAnimator(entry *donburi.Entry) (*Animator, bool) {
	if !entry.HasComponent(AnimatorComponent) {
		return nil, false
	}

	return AnimatorComponent.Get(entry), true
}

func (a *Animator) Animation() string {
	return a.animation
}

func (a *Animator) SetAnimation(animation string) {
	a.animation = animation
}

func (a *Animator) Library() LibraryID {
	return a.library
}

func (a *Animator) SetLibrary(library LibraryID) {
	a.library = library
}

func (a *Animator) Elapsed() float64 {
	return a.elapsed
}

func (a *Animator) SetElapsed(elapsed float64) {
	a.elapsed = elapsed
}

func (a *Animator) PingPongReversing() bool {
	return a.pingPongReversing
}

func (a *Animator) SetPingPongReversing(pingPongReversing bool) {
	a.pingPongReversing = pingPongReversing
}
