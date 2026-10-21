package game

type Update func(dt float64) error

type GameUpdate struct {
	time    *Time
	updates []Update
	fixed   []Update
	late    []Update
}

func NewGameUpdate(gt *Time) *GameUpdate {
	return &GameUpdate{
		time: gt,
	}
}

func (gu *GameUpdate) AddUpdate(update Update) {
	gu.updates = append(gu.updates, update)
}

func (gu *GameUpdate) AddFixed(update Update) {
	gu.fixed = append(gu.fixed, update)
}

func (gu *GameUpdate) AddLate(update Update) {
	gu.late = append(gu.late, update)
}

func (gu *GameUpdate) Update() error {
	gu.time.Tick()

	dt := gu.time.DeltaTime()
	fdt := gu.time.FixedDeltaTime()
	steps := gu.time.FixedSteps()

	if err := gu.callUpdates(dt, gu.updates); err != nil {
		return err
	}

	for range steps {
		if err := gu.callUpdates(fdt, gu.fixed); err != nil {
			return err
		}
	}

	if err := gu.callUpdates(dt, gu.late); err != nil {
		return err
	}

	return nil
}

func (gu *GameUpdate) callUpdates(dt float64, updates []Update) error {
	for _, update := range updates {
		if err := update(dt); err != nil {
			return err
		}
	}
	return nil
}
