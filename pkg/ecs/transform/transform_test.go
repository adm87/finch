package transform

import (
	"math"
	"testing"

	"github.com/yohamta/donburi"
)

func newEntry(w donburi.World) *donburi.Entry {
	return w.Entry(w.Create(Component))
}

func apply(t *Transform, x, y float64) (float64, float64) {
	m := t.Matrix()
	return m.Apply(x, y)
}

func near(a, b float64) bool {
	return math.Abs(a-b) < 1e-9
}

func TestSettersBumpVersionOnlyOnChange(t *testing.T) {
	tests := []struct {
		name string
		same func(*Transform)
		diff func(*Transform)
	}{
		{"position", func(t *Transform) { t.SetPosition(0, 0) }, func(t *Transform) { t.SetPosition(1, 2) }},
		{"scale", func(t *Transform) { t.SetScale(1, 1) }, func(t *Transform) { t.SetScale(2, 2) }},
		{"origin", func(t *Transform) { t.SetOrigin(0, 0) }, func(t *Transform) { t.SetOrigin(3, 4) }},
		{"angle", func(t *Transform) { t.SetAngle(0) }, func(t *Transform) { t.SetAngle(1) }},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tr := Component.Get(newEntry(donburi.NewWorld()))
			start := tr.Version()

			tt.same(tr)
			if tr.Version() != start {
				t.Errorf("setting the same value bumped the version to %d", tr.Version())
			}

			tt.diff(tr)
			if tr.Version() != start+1 {
				t.Errorf("version = %d, want %d", tr.Version(), start+1)
			}
		})
	}
}

func TestUpdateBuildsNewTransforms(t *testing.T) {
	w := donburi.NewWorld()
	tr := Component.Get(newEntry(w))
	tr.matrix.Translate(5, 5) // anything but identity

	Update(w)

	if x, y := apply(tr, 1, 1); !near(x, 1) || !near(y, 1) {
		t.Errorf("default matrix moved (1, 1) to (%v, %v), want identity", x, y)
	}
}

func TestUpdateOnlyRebuildsChangedTransforms(t *testing.T) {
	w := donburi.NewWorld()
	changed := Component.Get(newEntry(w))
	unchanged := Component.Get(newEntry(w))
	Update(w)

	changed.SetPosition(10, 0)
	unchanged.matrix.Translate(99, 0) // only visible if Update rebuilds it
	Update(w)

	if x, _ := apply(changed, 0, 0); !near(x, 10) {
		t.Errorf("changed transform moved (0, 0) to x = %v, want 10", x)
	}
	if x, _ := apply(unchanged, 0, 0); !near(x, 99) {
		t.Errorf("unchanged transform was rebuilt")
	}
}

func TestMatrixScalesAndRotatesAroundOrigin(t *testing.T) {
	tests := []struct {
		name         string
		set          func(*Transform)
		px, py       float64
		wantX, wantY float64
	}{
		{"position", func(t *Transform) { t.SetPosition(10, 20) }, 0, 0, 10, 20},
		{"origin stays put", func(t *Transform) {
			t.SetPosition(10, 20)
			t.SetOrigin(4, 4)
			t.SetScale(3, 3)
			t.SetAngle(math.Pi / 2)
		}, 4, 4, 14, 24},
		{"scale around origin", func(t *Transform) {
			t.SetOrigin(4, 4)
			t.SetScale(2, 2)
		}, 5, 4, 6, 4},
		{"rotate around origin", func(t *Transform) {
			t.SetOrigin(4, 4)
			t.SetAngle(math.Pi / 2)
		}, 5, 4, 4, 5},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := donburi.NewWorld()
			tr := Component.Get(newEntry(w))
			tt.set(tr)
			Update(w)

			if x, y := apply(tr, tt.px, tt.py); !near(x, tt.wantX) || !near(y, tt.wantY) {
				t.Errorf("(%v, %v) moved to (%v, %v), want (%v, %v)", tt.px, tt.py, x, y, tt.wantX, tt.wantY)
			}
		})
	}
}

func TestGet(t *testing.T) {
	w := donburi.NewWorld()
	if _, ok := Get(newEntry(w)); !ok {
		t.Error("Get on an entity with a Transform returned false")
	}
	if _, ok := Get(w.Entry(w.Create(donburi.NewTag()))); ok {
		t.Error("Get on an entity without a Transform returned true")
	}
}
