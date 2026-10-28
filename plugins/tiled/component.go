package tiled

import (
	"github.com/yohamta/donburi"
)

// TileMap draws a map created by Maps.
type TileMap struct {
	handle Handle
}

var TileMapComponent = donburi.NewComponentType[TileMap]()

func GetTileMap(entry *donburi.Entry) (*TileMap, bool) {
	if !entry.HasComponent(TileMapComponent) {
		return nil, false
	}

	return TileMapComponent.Get(entry), true
}

func (t *TileMap) Handle() Handle {
	return t.handle
}

func (t *TileMap) SetHandle(handle Handle) {
	t.handle = handle
}
