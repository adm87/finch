package game

import (
	"github.com/adm87/finch/cmd/game/internal/config"
	"github.com/adm87/finch/pkg/application"
	"github.com/adm87/finch/pkg/aseprite"
	"github.com/adm87/finch/pkg/rendering"
)

type Game struct {
	App       *application.Application
	Aseprite  *aseprite.Libraries
	Config    *config.GameConfig
	Renderers *Renderers
	Time      *Time
}

type Renderers struct {
	Image rendering.AdapterID
}
