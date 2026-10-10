package rendering

import (
	"github.com/adm87/finch/pkg/ecs/components/renderer"
	"github.com/yohamta/donburi"
)

type RenderCandidate struct {
	Entry    *donburi.Entry
	Renderer *renderer.Renderer
}

func (rc *RenderCandidate) IsValid() bool {
	return rc != nil && rc.Entry != nil && rc.Renderer != nil
}

func (rc *RenderCandidate) Clear() {
	rc.Entry = nil
	rc.Renderer = nil
}
