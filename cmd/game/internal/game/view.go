package game

import (
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/yohamta/donburi"

	"github.com/adm87/finch/pkg/ecs/components"
	"github.com/adm87/finch/pkg/geom"
)

type View struct {
	camera *donburi.Entry
	screen *Screen
}

func NewViewport(screen *Screen) *View {
	return &View{
		screen: screen,
	}
}

func (v *View) Camera() *donburi.Entry {
	return v.camera
}

func (v *View) SetCamera(camera *donburi.Entry) {
	v.camera = camera

	if v.camera != nil {
		if t, tOk := components.GetTransform(v.camera); tOk {
			v.refresh(t)
		}
	}
}

func (v *View) Viewport() (geom.Rect, ebiten.GeoM) {
	if v.camera == nil {
		return geom.Rect{}, ebiten.GeoM{}
	}

	t, tOk := components.GetTransform(v.camera)
	if !tOk {
		return geom.Rect{}, ebiten.GeoM{}
	}

	bounds := v.refresh(t)

	geoM := t.GeoM()
	if !geoM.IsInvertible() {
		return bounds, ebiten.GeoM{}
	}

	geoM.Invert()
	return bounds, geoM
}

func (v *View) WorldToScreen(worldX, worldY float64) (screenX, screenY float64) {
	if v.camera == nil {
		return worldX, worldY
	}

	_, geoM := v.Viewport()
	return geoM.Apply(worldX, worldY)
}

func (v *View) ScreenToWorld(screenX, screenY float64) (worldX, worldY float64) {
	if v.camera == nil {
		return screenX, screenY
	}

	t, tOk := components.GetTransform(v.camera)
	if !tOk {
		return screenX, screenY
	}

	geoM := t.GeoM()
	return geoM.Apply(v.screen.ScreenToSafeArea(screenX, screenY))
}

func (v *View) refresh(t *components.Transform) geom.Rect {
	safeArea := v.screen.SafeArea()
	t.SetOrigin(safeArea.Center())

	b, bOk := components.GetBounds(v.camera)
	if !bOk {
		return geom.Rect{}
	}
	b.SetRect(safeArea)

	return b.WorldRect(t)
}
