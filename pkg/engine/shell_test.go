package engine

import (
	"errors"
	"slices"
	"testing"
)

func TestShellUpdateRunsPhasesInOrder(t *testing.T) {
	var calls []string
	s := &shell{
		clock:        newClock(60, 120),
		updater:      recordUpdate(&calls, "update", nil),
		fixedUpdater: recordUpdate(&calls, "fixed", nil),
		laterUpdater: recordUpdate(&calls, "later", nil),
	}

	if err := s.Update(); err != nil {
		t.Fatalf("Update() = %v", err)
	}

	want := []string{"update", "fixed", "fixed", "later"}
	if !slices.Equal(calls, want) {
		t.Errorf("calls = %q, want %q", calls, want)
	}
}

func TestShellUpdateStopsOnError(t *testing.T) {
	errStop := errors.New("stop")

	tests := []struct {
		name                string
		update, fixed, late error
		want                []string
	}{
		{"update", errStop, nil, nil, []string{"update"}},
		{"fixed", nil, errStop, nil, []string{"update", "fixed"}},
		{"later", nil, nil, errStop, []string{"update", "fixed", "fixed", "later"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var calls []string
			s := &shell{
				clock:        newClock(60, 120),
				updater:      recordUpdate(&calls, "update", tt.update),
				fixedUpdater: recordUpdate(&calls, "fixed", tt.fixed),
				laterUpdater: recordUpdate(&calls, "later", tt.late),
			}

			if err := s.Update(); !errors.Is(err, errStop) {
				t.Errorf("Update() = %v, want %v", err, errStop)
			}
			if !slices.Equal(calls, tt.want) {
				t.Errorf("calls = %q, want %q", calls, tt.want)
			}
		})
	}
}

func TestShellLayoutUsesDisplay(t *testing.T) {
	s := &shell{display: DisplayFunc(func(w, h float64) (float64, float64) {
		return w * 2, h * 2
	})}

	if w, h := s.LayoutF(100, 50); w != 200 || h != 100 {
		t.Errorf("LayoutF() = (%v, %v), want (200, 100)", w, h)
	}
	if w, h := s.Layout(100, 50); w != 200 || h != 100 {
		t.Errorf("Layout() = (%v, %v), want (200, 100)", w, h)
	}
}

func recordUpdate(calls *[]string, name string, err error) Updater {
	return UpdaterFunc(func(float64) error {
		*calls = append(*calls, name)
		return err
	})
}
