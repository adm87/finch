package rendering

import (
	"github.com/adm87/finch/pkg/geom"
	"github.com/hajimehoshi/ebiten/v2"
)

type View interface {
	Viewport() (geom.Rect, ebiten.GeoM)
}

type RenderPipeline struct {
	view View
}

func NewRenderPipeline(viewport View) *RenderPipeline {
	return &RenderPipeline{
		view: viewport,
	}
}

func (rp *RenderPipeline) Draw(target *ebiten.Image) error {
	return nil
}
