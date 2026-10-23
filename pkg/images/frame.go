package images

import (
	"image"

	"github.com/hajimehoshi/ebiten/v2"
)

// Frame is a cut of a cached image. Offset and Size place a trimmed frame on its original canvas.
type Frame struct {
	Image  *ebiten.Image
	Offset image.Point
	Size   image.Point
}

// FrameRect describes how to cut a Frame. A zero Size means the size of Rect.
type FrameRect struct {
	Rect   image.Rectangle
	Offset image.Point
	Size   image.Point
}

func untrimmed(img *ebiten.Image) Frame {
	return Frame{Image: img, Size: img.Bounds().Size()}
}
