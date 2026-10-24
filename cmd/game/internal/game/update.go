package game

import "github.com/adm87/finch/pkg/application"

type UpdateFunc func(dt float64) error

type Updater struct {
	time    *Time
	updates []UpdateFunc
	fixed   []UpdateFunc
	late    []UpdateFunc
}

func NewUpdater(gt *Time) *Updater {
	return &Updater{
		time: gt,
	}
}

func (u *Updater) AddUpdate(update UpdateFunc) {
	u.updates = append(u.updates, update)
}

func (u *Updater) AddFixed(update UpdateFunc) {
	u.fixed = append(u.fixed, update)
}

func (u *Updater) AddLate(update UpdateFunc) {
	u.late = append(u.late, update)
}

func (u *Updater) Update(*application.Application) error {
	u.time.Tick()

	dt := u.time.DeltaTime()
	fdt := u.time.FixedDeltaTime()
	steps := u.time.FixedSteps()

	if err := u.callUpdates(dt, u.updates); err != nil {
		return err
	}

	for range steps {
		if err := u.callUpdates(fdt, u.fixed); err != nil {
			return err
		}
	}

	if err := u.callUpdates(dt, u.late); err != nil {
		return err
	}

	return nil
}

func (u *Updater) callUpdates(dt float64, updates []UpdateFunc) error {
	for _, update := range updates {
		if err := update(dt); err != nil {
			return err
		}
	}
	return nil
}
