package game

import (
	"github.com/adm87/finch/cmd/game/internal/config"
	"github.com/adm87/finch/pkg/aseprite"
	"github.com/adm87/finch/pkg/resources"
	"github.com/adm87/finch/pkg/storage"
)

type Game struct {
	Aseprite  aseprite.Aseprite
	Resources resources.Resources
	Config    *config.GameConfig
	Cache     *GameCache
	Renderers *GameRenderers
	Screen    *Screen
	Time      *Time
	View      *View
}

type GameCache struct {
	Data   *storage.DataCache
	Images *storage.ImageCache
}

type GameRenderers struct {
	Image uint64
}
