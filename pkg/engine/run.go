package engine

import (
	"time"

	"github.com/hajimehoshi/ebiten/v2"
)

func Run(opts ...Option) error {
	o := defaultOptions()
	for _, opt := range opts {
		opt(&o)
	}

	ebiten.SetFullscreen(true)

	return ebiten.RunGame(&shell{
		clock:        clock{fixedDT: time.Second / time.Duration(max(60, 1))},
		display:      o.display,
		drawer:       o.drawer,
		updater:      o.updater,
		fixedUpdater: o.fixedUpdater,
		laterUpdater: o.laterUpdater,
	})
}
