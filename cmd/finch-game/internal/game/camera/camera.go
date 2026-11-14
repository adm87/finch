package camera

import (
	"math"

	"github.com/yohamta/donburi"

	"github.com/adm87/finch/pkg/display"
	"github.com/adm87/finch/pkg/ecs/bounds"
	"github.com/adm87/finch/pkg/ecs/rendering"
	"github.com/adm87/finch/pkg/ecs/transform"
	"github.com/adm87/finch/pkg/geom"
)

const smoothing = 10 // higher catches up faster

// Camera is where the camera is heading. Its transform eases towards it.
type Camera struct {
	targetX float64
	targetY float64
}

var Component = donburi.NewComponentType[Camera]()

// Main tags the camera the world is drawn through.
var Main = donburi.NewTag("camera.Main")

func (c *Camera) Target() (x, y float64) {
	return c.targetX, c.targetY
}

func (c *Camera) SetTarget(x, y float64) {
	c.targetX = x
	c.targetY = y
}

// New creates the main camera at x, y.
func New(w donburi.World, x, y float64) *donburi.Entry {
	e := w.Entry(w.Create(transform.Component, bounds.Component, Component, Main))
	transform.Component.Get(e).SetPosition(x, y)
	Component.Get(e).SetTarget(x, y)
	return e
}

// Fit centres the camera on its position and sizes its view to the screen.
func Fit(w donburi.World, screen *display.Screen) {
	e, ok := Main.First(w)
	if !ok {
		return
	}
	visible := screen.Visible()
	transform.Component.Get(e).SetOrigin(visible.Center())
	bounds.Component.Get(e).SetLocal(visible)
}

// Move shifts where the main camera is heading by dx, dy.
func Move(w donburi.World, dx, dy float64) {
	e, ok := Main.First(w)
	if !ok {
		return
	}
	c := Component.Get(e)
	x, y := c.Target()
	c.SetTarget(x+dx, y+dy)
}

// Ease moves the main camera part of the way to its target. The fraction
// comes from dt, so it eases at the same speed at any tick rate.
func Ease(w donburi.World, dt float64) {
	e, ok := Main.First(w)
	if !ok {
		return
	}
	t := transform.Component.Get(e)
	x, y := t.Position()
	tx, ty := Component.Get(e).Target()
	k := 1 - math.Exp(-smoothing*dt)
	t.SetPosition(x+(tx-x)*k, y+(ty-y)*k)
}

// View is what the main camera sees, or the whole screen if there's no camera.
func View(w donburi.World, screen geom.Rect) rendering.View {
	e, ok := Main.First(w)
	if !ok {
		return rendering.View{Bounds: screen}
	}
	m := transform.Component.Get(e).Matrix()
	view := rendering.View{Bounds: bounds.Component.Get(e).World(m), Matrix: m}
	view.Matrix.Invert()
	return view
}
