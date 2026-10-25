package debug

import (
	"github.com/yohamta/donburi"
)

type OverlayType uint32

const (
	OverlayTypeEntities OverlayType = iota
	OverlayTypeFPS
	OverlayTypePartitions
	OverlayTypeZeroZero
)

type Debug struct {
	Overlays uint32
}

var DebugComponent = donburi.NewComponentType[Debug]()

func GetDebug(ecs donburi.World) *Debug {
	entry, ok := DebugComponent.First(ecs)
	if !ok {
		return &Debug{}
	}
	return DebugComponent.Get(entry)
}

func (d *Debug) SetOverlay(overlay OverlayType, enabled bool) {
	if enabled {
		d.Overlays |= 1 << overlay
	} else {
		d.Overlays &^= 1 << overlay
	}
}

func (d *Debug) HasOverlay(overlay OverlayType) bool {
	return d.Overlays&(1<<overlay) != 0
}
