package rendering

import (
	"github.com/adm87/finch/pkg/ecs/components"
	"github.com/yohamta/donburi"
)

type RenderCandidate struct {
	Entry     *donburi.Entry
	Renderer  *components.Renderer
	Transform *components.Transform
	Bounds    *components.Bounds
}

func (rc *RenderCandidate) IsValid() bool {
	return rc != nil &&
		rc.Entry != nil &&
		rc.Renderer != nil &&
		rc.Transform != nil &&
		rc.Bounds != nil
}

func (rc *RenderCandidate) Clear() {
	rc.Entry = nil
	rc.Renderer = nil
	rc.Transform = nil
	rc.Bounds = nil
}
