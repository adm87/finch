package engine

import (
	"github.com/hajimehoshi/ebiten/v2"

	"github.com/adm87/finch/pkg/application"
)

type Updater interface {
	Update(app *application.Application) error
}

type Drawer interface {
	Draw(target *ebiten.Image, app *application.Application) error
}

type UpdaterFunc func(app *application.Application) error

func (f UpdaterFunc) Update(app *application.Application) error {
	return f(app)
}

type DrawerFunc func(target *ebiten.Image, app *application.Application) error

func (f DrawerFunc) Draw(target *ebiten.Image, app *application.Application) error {
	return f(target, app)
}

func noopUpdate(*application.Application) error {
	return nil
}

func noopDraw(*ebiten.Image, *application.Application) error {
	return nil
}
