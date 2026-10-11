package rendering

import (
	"github.com/adm87/finch/pkg/ecs/components/renderer"
	"github.com/adm87/finch/pkg/ecs/components/transform"
	"github.com/yohamta/donburi"
)

type RenderCandidate struct {
	Entry     *donburi.Entry
	Renderer  *renderer.Renderer
	Transform *transform.Transform
	Matrix    *transform.Matrix
}

func (rc *RenderCandidate) IsValid() bool {
	return rc != nil && rc.Entry != nil && rc.Renderer != nil && rc.Transform != nil && rc.Matrix != nil
}

func (rc *RenderCandidate) Clear() {
	rc.Entry = nil
	rc.Renderer = nil
	rc.Transform = nil
	rc.Matrix = nil
}
