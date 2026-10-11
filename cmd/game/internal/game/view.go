package game

import (
	"github.com/adm87/finch/pkg/ecs/components/transform"
	"github.com/adm87/finch/pkg/geom"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/yohamta/donburi"
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
}

func (v *View) Viewport() (geom.Rect, ebiten.GeoM) {
	if v.camera == nil {
		return geom.Rect{}, ebiten.GeoM{}
	}

	t, tOk := transform.GetTransform(v.camera)
	if !tOk {
		return geom.Rect{}, ebiten.GeoM{}
	}

	m, mOk := transform.GetMatrix(v.camera)
	if !mOk {
		return geom.Rect{}, ebiten.GeoM{}
	}

	safeArea := v.screen.SafeArea()
	t.SetOrigin(safeArea.Center())

	bounds := transform.TransformRect(t, m, safeArea)
	transform.SetRect(v.camera, bounds)

	geoM := transform.GetGeoMOf(t, m)
	if geoM.IsInvertible() {
		geoM.Invert()
	}

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

	t, tOk := transform.GetTransform(v.camera)
	if !tOk {
		return screenX, screenY
	}

	m, mOk := transform.GetMatrix(v.camera)
	if !mOk {
		return screenX, screenY
	}

	geoM := transform.GetGeoMOf(t, m)
	return geoM.Apply(v.screen.ScreenToSafeArea(screenX, screenY))
}
