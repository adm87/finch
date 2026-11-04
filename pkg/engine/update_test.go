package engine

import (
	"errors"
	"slices"
	"testing"
)

func TestUpdatePhaseRunsInOrder(t *testing.T) {
	var calls []string
	var dts []float64
	record := func(name string) Updater {
		return UpdaterFunc(func(dt float64) error {
			calls = append(calls, name)
			dts = append(dts, dt)
			return nil
		})
	}

	phase := UpdatePhase{record("a"), record("b"), record("c")}
	if err := phase.Update(0.5); err != nil {
		t.Fatalf("Update() = %v", err)
	}

	if want := []string{"a", "b", "c"}; !slices.Equal(calls, want) {
		t.Errorf("calls = %q, want %q", calls, want)
	}
	if want := []float64{0.5, 0.5, 0.5}; !slices.Equal(dts, want) {
		t.Errorf("dt = %v, want %v", dts, want)
	}
}

func TestUpdatePhaseStopsAtFirstError(t *testing.T) {
	errStop := errors.New("stop")
	var calls []string

	phase := UpdatePhase{
		recordUpdate(&calls, "a", nil),
		recordUpdate(&calls, "b", errStop),
		recordUpdate(&calls, "c", nil),
	}
	if err := phase.Update(1); !errors.Is(err, errStop) {
		t.Errorf("Update() = %v, want %v", err, errStop)
	}
	if want := []string{"a", "b"}; !slices.Equal(calls, want) {
		t.Errorf("calls = %q, want %q", calls, want)
	}
}

func TestEmptyUpdatePhase(t *testing.T) {
	if err := (UpdatePhase{}).Update(1); err != nil {
		t.Errorf("Update() = %v, want nil", err)
	}
}
