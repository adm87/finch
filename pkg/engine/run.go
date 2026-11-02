package engine

import (
	"time"

	"github.com/hajimehoshi/ebiten/v2"
)

func Run(opts ...Option) error {
	options := DefaultOptions()
	for _, opt := range opts {
		opt(&options)
	}

	return ebiten.RunGame(&shell{
		clock:        clock{fixedDT: time.Second / time.Duration(max(60, 1))},
		display:      options.display,
		drawer:       options.drawer,
		updater:      options.updater,
		fixedUpdater: options.fixedUpdater,
		laterUpdater: options.laterUpdater,
	})
}
