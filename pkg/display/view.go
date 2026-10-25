package display

import (
	"github.com/yohamta/donburi"

	"github.com/adm87/finch/pkg/geom"
	"github.com/adm87/finch/pkg/transform"
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

func (v *View) HasCamera() bool {
	return v.camera != nil
}

func (v *View) SetCamera(camera *donburi.Entry) {
	v.camera = camera

	if v.camera != nil {
		if t, tOk := transform.GetTransform(v.camera); tOk {
			v.refresh(t)
		}
	}
}

func (v *View) Viewport() (geom.Rect, geom.Matrix) {
	if v.camera == nil {
		return geom.Rect{}, geom.Matrix{}
	}

	t, tOk := transform.GetTransform(v.camera)
	if !tOk {
		return geom.Rect{}, geom.Matrix{}
	}

	bounds := v.refresh(t)

	geoM := t.GeoM()
	if !geoM.IsInvertible() {
		return bounds, geom.Matrix{}
	}

	geoM.Invert()
	return bounds, geoM
}

func (v *View) WorldToScreen(worldX, worldY float64) (screenX, screenY float64) {
	if v.camera == nil {
		return worldX, worldY
	}

	_, geoM := v.Viewport()
	return v.screen.SafeAreaToScreen(geoM.Apply(worldX, worldY))
}

func (v *View) ScreenToWorld(screenX, screenY float64) (worldX, worldY float64) {
	if v.camera == nil {
		return screenX, screenY
	}

	t, tOk := transform.GetTransform(v.camera)
	if !tOk {
		return screenX, screenY
	}

	v.refresh(t)
	geoM := t.GeoM()
	return geoM.Apply(v.screen.ScreenToSafeArea(screenX, screenY))
}

func (v *View) refresh(t *transform.Transform) geom.Rect {
	safeArea := v.screen.SafeArea()
	t.SetOrigin(safeArea.Center())

	b, bOk := transform.GetBounds(v.camera)
	if !bOk {
		return geom.Rect{}
	}
	b.SetRect(safeArea)

	return b.WorldRect(t)
}
