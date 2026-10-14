package game

import (
	"github.com/adm87/finch/cmd/game/internal/config"
	"github.com/adm87/finch/pkg/images"
	"github.com/adm87/finch/pkg/resources"
	"github.com/adm87/finch/pkg/resources/data"
)

type Game struct {
	Resources resources.Resources
	Config    *config.GameConfig
	Cache     *GameCache
	Renderers *GameRenderers
}

type GameCache struct {
	Data   *data.DataCache
	Images *images.ImageCache
}

type GameRenderers struct {
	Image uint64
}
