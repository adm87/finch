package rendering

import (
	"github.com/yohamta/donburi"

	"github.com/adm87/finch/pkg/transform"
)

type Candidate struct {
	Entry     *donburi.Entry
	Renderer  *Renderer
	Transform *transform.Transform
	Bounds    *transform.Bounds
}

func (rc *Candidate) IsValid() bool {
	return rc != nil &&
		rc.Entry != nil &&
		rc.Renderer != nil &&
		rc.Transform != nil &&
		rc.Bounds != nil
}

func (rc *Candidate) Clear() {
	rc.Entry = nil
	rc.Renderer = nil
	rc.Transform = nil
	rc.Bounds = nil
}
