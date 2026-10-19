package components

import "github.com/yohamta/donburi"

type Aseprite struct {
	animation string
	library   uint64
}

var AsepriteComponent = donburi.NewComponentType[Aseprite]()

func GetAseprite(entry *donburi.Entry) (*Aseprite, bool) {
	if !entry.HasComponent(AsepriteComponent) {
		return nil, false
	}

	return AsepriteComponent.Get(entry), true
}

func (a *Aseprite) Animation() string {
	return a.animation
}

func (a *Aseprite) SetAnimation(animation string) {
	a.animation = animation
}

func (a *Aseprite) Library() uint64 {
	return a.library
}

func (a *Aseprite) SetLibrary(library uint64) {
	a.library = library
}
