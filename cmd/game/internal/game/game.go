package game

import (
	"github.com/adm87/finch/cmd/game/internal/config"
	"github.com/adm87/finch/pkg/aseprite"
	"github.com/adm87/finch/pkg/assets"
	"github.com/adm87/finch/pkg/data"
	"github.com/adm87/finch/pkg/display"
	"github.com/adm87/finch/pkg/images"
	"github.com/adm87/finch/pkg/rendering"
)

type Game struct {
	Aseprite  *aseprite.Libraries
	Assets    *assets.Assets
	Config    *config.GameConfig
	Cache     *Cache
	Renderers *Renderers
	Screen    *display.Screen
	View      *display.View
	Time      *Time
}

type Cache struct {
	Data   *data.Cache
	Images *images.Cache
}

type Renderers struct {
	Image rendering.AdapterID
}
