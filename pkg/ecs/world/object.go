package world

import (
	"math"

	"github.com/adm87/finch/pkg/ecs/components/bounds"
	"github.com/adm87/finch/pkg/ecs/components/matrix"
	"github.com/adm87/finch/pkg/ecs/components/transform"
	"github.com/adm87/finch/pkg/geom"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/yohamta/donburi"
)

var baseComponents = []donburi.IComponentType{
	bounds.ComponentType,
	matrix.ComponentType,
	transform.ComponentType,
	index,
}

// WorldObject represents an entity in the ECS world with common components like bounds, matrix, and transform.
type WorldObject donburi.Entity

// NewWorldObject creates a new WorldObject in the given ECS world with the specified components.
// It automatically includes the base components like bounds, matrix, and transform.
func NewWorldObject(world donburi.World, components ...donburi.IComponentType) WorldObject {
	if len(components) == 0 {
		return WorldObject(world.Create(baseComponents...))
	}

	comps := make([]donburi.IComponentType, 0, len(baseComponents)+len(components))
	comps = append(comps, baseComponents...)
	comps = append(comps, components...)

	return WorldObject(world.Create(comps...))
}

// IsValid checks that the object exists and has all base components.
func (obj WorldObject) IsValid(world donburi.World) bool {
	if !world.Valid(obj.Entity()) {
		return false
	}

	entry := obj.Entry(world)
	for _, c := range baseComponents {
		if !entry.HasComponent(c) {
			return false
		}
	}

	return true
}

func (obj WorldObject) Entity() donburi.Entity { return donburi.Entity(obj) }

func (obj WorldObject) Entry(world donburi.World) *donburi.Entry {
	return world.Entry(obj.Entity())
}

// The accessors below rely on NewWorldObject guaranteeing the base components.
// Use IsValid at boundaries where that isn't certain.
func (obj WorldObject) Transform(world donburi.World) *transform.Transform {
	return transform.ComponentType.Get(obj.Entry(world))
}

func (obj WorldObject) Bounds(world donburi.World) *bounds.Bounds {
	return bounds.ComponentType.Get(obj.Entry(world))
}

func (obj WorldObject) GeoM(world donburi.World) ebiten.GeoM {
	return GeoMOf(obj.Entry(world))
}

func (obj WorldObject) LocalBounds(world donburi.World) geom.Rect {
	return obj.Bounds(world).Rect()
}

// WorldBounds returns the local bounds transformed by the current GeoM.
// Note: this refreshes the matrix and clears the transform's dirty flag.
func (obj WorldObject) WorldBounds(world donburi.World) geom.Rect {
	return WorldBoundsOf(obj.Entry(world))
}

func GeoMOf(entry *donburi.Entry) ebiten.GeoM {
	t := transform.ComponentType.Get(entry)
	m := matrix.ComponentType.Get(entry)

	if t.IsDirty() {
		x, y := t.Position()
		sx, sy := t.Scale()
		ox, oy := t.Origin()
		rot := t.Rotation()

		m.Update(x, y, sx, sy, ox, oy, rot)
		t.SetDirty(false)
	}

	return m.GeoM()
}

func WorldBoundsOf(entry *donburi.Entry) geom.Rect {
	b := bounds.ComponentType.Get(entry).Rect()
	m := GeoMOf(entry)

	x1, y1 := b.Min()
	x2, y2 := b.Max()

	minX, minY := math.Inf(1), math.Inf(1)
	maxX, maxY := math.Inf(-1), math.Inf(-1)

	for _, c := range [4][2]float64{{x1, y1}, {x2, y1}, {x2, y2}, {x1, y2}} {
		x, y := m.Apply(c[0], c[1])
		minX, maxX = min(minX, x), max(maxX, x)
		minY, maxY = min(minY, y), max(maxY, y)
	}

	return geom.NewRect(minX, minY, maxX-minX, maxY-minY)
}
