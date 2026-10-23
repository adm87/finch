package game

import (
	"github.com/adm87/finch/cmd/game/internal/config"
	"github.com/adm87/finch/pkg/aseprite"
	"github.com/adm87/finch/pkg/assets"
	"github.com/adm87/finch/pkg/data"
	"github.com/adm87/finch/pkg/images"
)

type Game struct {
	Aseprite  aseprite.Libraries
	Assets    *assets.Assets
	Config    *config.GameConfig
	Cache     *GameCache
	Renderers *GameRenderers
	Screen    *Screen
	Time      *Time
	View      *View
}

type GameCache struct {
	Data   *data.Cache
	Images *images.Cache
}

type GameRenderers struct {
	Image uint64
}
