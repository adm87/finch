package engine

import (
	"errors"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/yohamta/donburi"

	"github.com/adm87/finch/pkg/application"
	"github.com/adm87/finch/pkg/geom"
)

// System runs once per update phase, with dt in seconds.
type System func(app *application.Application, dt float64) error

// Startup runs once, before the first update.
type Startup func(app *application.Application) error

// DrawContext holds what a drawer needs for one frame. It's reused, so don't hold on to it.
type DrawContext struct {
	Entries []*donburi.Entry
	View    geom.Rect
	Matrix  geom.Matrix
}

type Drawer interface {
	Draw(target *ebiten.Image, ctx *DrawContext) error
}

type DrawerFunc func(target *ebiten.Image, ctx *DrawContext) error

func (f DrawerFunc) Draw(target *ebiten.Image, ctx *DrawContext) error {
	return f(target, ctx)
}

type Drawers []Drawer

func (d Drawers) Draw(target *ebiten.Image, ctx *DrawContext) error {
	var errList []error
	for _, drawer := range d {
		if err := drawer.Draw(target, ctx); err != nil {
			errList = append(errList, err)
		}
	}
	return errors.Join(errList...)
}
