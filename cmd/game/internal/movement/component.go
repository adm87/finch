package movement

import "github.com/yohamta/donburi"

type Movement struct {
	dx    float64
	dy    float64
	speed float64
}

var MovementComponent = donburi.NewComponentType[Movement]()

func GetMovement(entry *donburi.Entry) (*Movement, bool) {
	if entry.HasComponent(MovementComponent) {
		return MovementComponent.Get(entry), true
	}
	return nil, false
}

func (m *Movement) Delta() (float64, float64) {
	return m.dx, m.dy
}

func (m *Movement) SetDelta(dx, dy float64) {
	m.dx = dx
	m.dy = dy
}

func (m *Movement) Speed() float64 {
	return m.speed
}

func (m *Movement) SetSpeed(speed float64) {
	m.speed = speed
}
