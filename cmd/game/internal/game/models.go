package game

import (
	"github.com/adm87/finch/pkg/images"
	"github.com/adm87/finch/pkg/resources"
	"github.com/adm87/finch/pkg/resources/data"
)

type GameResources struct {
	Resources resources.Resources
	Cache     *GameCache
}

type GameCache struct {
	Data   *data.DataCache
	Images *images.ImageCache
}

type GameRenderers struct {
	Images uint64
}
