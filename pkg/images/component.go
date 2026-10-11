package images

import "github.com/yohamta/donburi"

type Image struct {
	ax, ay float64
}

var ImageComponent = donburi.NewComponentType[Image]()

func GetImage(entry *donburi.Entry) (*Image, bool) {
	if !entry.HasComponent(ImageComponent) {
		return nil, false
	}

	return ImageComponent.Get(entry), true
}

func (i *Image) Anchor() (float64, float64) {
	return i.ax, i.ay
}

func (i *Image) SetAnchor(ax, ay float64) {
	i.ax = ax
	i.ay = ay
}
