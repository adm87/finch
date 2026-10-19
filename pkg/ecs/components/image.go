package components

import (
	"image/color"

	"github.com/yohamta/donburi"
)

type Image struct {
	handle uint64
	frame  int
	ax, ay float64
	color  color.RGBA
}

var ImageComponent = donburi.NewComponentType[Image](Image{
	color: color.RGBA{R: 255, G: 255, B: 255, A: 255},
})

func GetImage(entry *donburi.Entry) (*Image, bool) {
	if !entry.HasComponent(ImageComponent) {
		return nil, false
	}

	return ImageComponent.Get(entry), true
}

func (i *Image) Handle() uint64 {
	return i.handle
}

func (i *Image) SetHandle(handle uint64) {
	i.handle = handle
}

func (i *Image) Frame() int {
	return i.frame
}

func (i *Image) SetFrame(frame int) {
	i.frame = frame
}

func (i *Image) Anchor() (float64, float64) {
	return i.ax, i.ay
}

func (i *Image) SetAnchor(ax, ay float64) {
	i.ax = ax
	i.ay = ay
}

func (i *Image) Color() color.RGBA {
	return i.color
}

func (i *Image) SetColor(c color.RGBA) {
	i.color = c
}
