package game

import (
	"github.com/adm87/finch/pkg/images"
	"github.com/adm87/finch/pkg/resources/data"
)

type GameCache struct {
	Data   *data.DataCache
	Images *images.ImageCache
}

type GameRenderers struct {
	Images uint64
}
