package engine

import (
	"testing"
)

func TestDefaultOptionsAreNoops(t *testing.T) {
	o := defaultOptions()

	if w, h := o.display.Layout(320, 180); w != 320 || h != 180 {
		t.Errorf("default Layout() = (%v, %v), want (320, 180)", w, h)
	}
	o.drawer.Draw(nil)
	for _, u := range []Updater{o.updater, o.fixedUpdater, o.laterUpdater} {
		if err := u.Update(1); err != nil {
			t.Errorf("default Update() = %v, want nil", err)
		}
	}
}

func TestOptionsRejectNil(t *testing.T) {
	tests := []struct {
		name   string
		option func() Option
	}{
		{"WithDisplay", func() Option { return WithDisplay(nil) }},
		{"WithDrawer", func() Option { return WithDrawer(nil) }},
		{"WithUpdater", func() Option { return WithUpdater(nil) }},
		{"WithFixedUpdater", func() Option { return WithFixedUpdater(nil) }},
		{"WithLaterUpdater", func() Option { return WithLaterUpdater(nil) }},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defer func() {
				if recover() == nil {
					t.Error("did not panic")
				}
			}()
			tt.option()
		})
	}
}
