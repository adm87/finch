package rendering

import (
	"github.com/yohamta/donburi"

	"github.com/adm87/finch/pkg/geom"
	"github.com/adm87/finch/pkg/transform"
)

// Job is a single rendering operation provided by an Adapter or rendering candidate.
type Job struct {
	Entry     *donburi.Entry
	Transform *transform.Transform
	adapter   Adapter
	Bounds    geom.Rect // world space, jobs outside the view are skipped
	Pass      int
	Depth     int32
	Layer     int16
}
