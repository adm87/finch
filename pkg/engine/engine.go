package engine

import (
	"errors"

	"github.com/hajimehoshi/ebiten/v2"

	"github.com/adm87/finch/pkg/application"
)

// System runs once per update phase. dt is in seconds.
type System func(app *application.Application, dt float64) error

type Drawer interface {
	Draw(target *ebiten.Image, app *application.Application) error
}

type DrawerFunc func(target *ebiten.Image, app *application.Application) error

func (f DrawerFunc) Draw(target *ebiten.Image, app *application.Application) error {
	return f(target, app)
}

type Drawers []Drawer

func (d Drawers) Draw(target *ebiten.Image, app *application.Application) error {
	var errList []error
	for _, drawer := range d {
		if err := drawer.Draw(target, app); err != nil {
			errList = append(errList, err)
		}
	}
	return errors.Join(errList...)
}
