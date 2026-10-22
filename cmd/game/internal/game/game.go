package game

import (
	"github.com/adm87/finch/cmd/game/internal/config"
	"github.com/adm87/finch/pkg/aseprite"
	"github.com/adm87/finch/pkg/assets"
	"github.com/adm87/finch/pkg/cache"
)

type Game struct {
	Aseprite  aseprite.Aseprite
	Assets    *assets.Assets
	Config    *config.GameConfig
	Cache     *GameCache
	Renderers *GameRenderers
	Screen    *Screen
	Time      *Time
	View      *View
}

type GameCache struct {
	Data   *cache.DataCache
	Images *cache.ImageCache
}

type GameRenderers struct {
	Image uint64
}
