// Package shuffle is a throwaway movement for test entities: each one loops
// around its home point so there's something moving under the effects.
package shuffle

import (
	"math"

	"github.com/yohamta/donburi"

	"github.com/adm87/finch/pkg/ecs/transform"
)

type Shuffle struct {
	homeX, homeY float64
	radius       float64
	speed        float64 // radians per second
	phase        float64
	elapsed      float64
}

var Component = donburi.NewComponentType[Shuffle]()

func (s *Shuffle) Home() (x, y float64) {
	return s.homeX, s.homeY
}

func (s *Shuffle) Radius() float64 {
	return s.radius
}

func (s *Shuffle) Speed() float64 {
	return s.speed
}

func (s *Shuffle) Phase() float64 {
	return s.phase
}

func (s *Shuffle) SetHome(x, y float64) {
	s.homeX = x
	s.homeY = y
}

func (s *Shuffle) SetRadius(radius float64) {
	s.radius = radius
}

func (s *Shuffle) SetSpeed(speed float64) {
	s.speed = speed
}

func (s *Shuffle) SetPhase(phase float64) {
	s.phase = phase
}

// Update moves every shuffling entity along a figure-of-eight-ish loop around its home.
func Update(w donburi.World, dt float64) {
	for e := range Component.Iter(w) {
		if !e.HasComponent(transform.Component) {
			continue
		}
		s := Component.Get(e)
		s.elapsed += dt

		a := s.elapsed*s.speed + s.phase
		transform.Component.Get(e).SetPosition(
			s.homeX+math.Cos(a)*s.radius,
			s.homeY+math.Sin(a*1.7)*s.radius*0.6,
		)
	}
}
