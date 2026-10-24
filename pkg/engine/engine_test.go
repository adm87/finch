package engine

import (
	"errors"
	"testing"
	"time"

	"github.com/go-openapi/testify/require"
	"github.com/hajimehoshi/ebiten/v2"

	"github.com/adm87/finch/pkg/application"
	"github.com/adm87/finch/pkg/logger"
)

type fakeTime struct{ t time.Time }

func (f *fakeTime) now() time.Time { return f.t }

func (f *fakeTime) advance(d time.Duration) { f.t = f.t.Add(d) }

func newFakeClock(tickRate int) (*clock, *fakeTime) {
	ft := &fakeTime{t: time.Unix(1000, 0)}
	return newClock(tickRate, ft.now), ft
}

func TestClockFirstTickHasNoDelta(t *testing.T) {
	c, _ := newFakeClock(60)
	c.tick()
	require.Zero(t, c.delta)
	require.Zero(t, c.fixedSteps)
}

func TestClockAccumulatesFixedSteps(t *testing.T) {
	c, ft := newFakeClock(10)
	c.tick()

	ft.advance(60 * time.Millisecond)
	c.tick()
	require.Equal(t, 60*time.Millisecond, c.delta)
	require.Zero(t, c.fixedSteps)

	ft.advance(60 * time.Millisecond)
	c.tick()
	require.Equal(t, 1, c.fixedSteps)

	ft.advance(250 * time.Millisecond)
	c.tick()
	require.Equal(t, 2, c.fixedSteps)
}

func TestClockCapsLongAndBackwardsGaps(t *testing.T) {
	c, ft := newFakeClock(10)
	c.tick()

	ft.advance(time.Hour)
	c.tick()
	require.Equal(t, maxDelta, c.delta)
	require.Equal(t, 10, c.fixedSteps)

	ft.advance(-time.Minute)
	c.tick()
	require.Zero(t, c.delta)
	require.Zero(t, c.fixedSteps)
}

func TestShellRunsPhasesInOrder(t *testing.T) {
	app, err := application.New(nil, logger.N)
	require.NoError(t, err)

	var calls []string
	record := func(name string) System {
		return func(_ *application.Application, dt float64) error {
			calls = append(calls, name)
			require.Positive(t, dt)
			return nil
		}
	}

	options := DefaultRunOptions()
	WithUpdate(record("update"))(options)
	WithFixedUpdate(record("fixed"))(options)
	WithLateUpdate(record("late"), record("late 2"))(options)

	c, ft := newFakeClock(10)
	sh := newShell(app, options, c)
	c.tick()
	ft.advance(200 * time.Millisecond)

	require.NoError(t, sh.Update())
	require.Equal(t, []string{"update", "fixed", "fixed", "late", "late 2"}, calls)
}

func TestShellStopsAtFirstError(t *testing.T) {
	app, err := application.New(nil, logger.N)
	require.NoError(t, err)

	stop := errors.New("stop")
	late := false

	options := DefaultRunOptions()
	WithUpdate(func(*application.Application, float64) error { return stop })(options)
	WithLateUpdate(func(*application.Application, float64) error { late = true; return nil })(options)

	c, _ := newFakeClock(60)
	require.ErrorIs(t, newShell(app, options, c).Update(), stop)
	require.False(t, late)
}

func TestDrawersDrawInOrderAndReportEveryError(t *testing.T) {
	var calls []string
	errA, errB := errors.New("a failed"), errors.New("b failed")
	drawer := func(name string, err error) Drawer {
		return DrawerFunc(func(*ebiten.Image, *application.Application) error {
			calls = append(calls, name)
			return err
		})
	}

	err := Drawers{drawer("a", errA), drawer("b", nil), drawer("c", errB)}.Draw(nil, nil)

	require.Equal(t, []string{"a", "b", "c"}, calls)
	require.ErrorIs(t, err, errA)
	require.ErrorIs(t, err, errB)
	require.NoError(t, Drawers(nil).Draw(nil, nil))
}

func TestShellDefaultsWorldDrawersWithoutChangingOptions(t *testing.T) {
	app, err := application.New(nil, logger.N)
	require.NoError(t, err)

	options := DefaultRunOptions()
	c, _ := newFakeClock(60)
	sh := newShell(app, options, c)

	require.Len(t, sh.worldDrawers, 1)
	require.Empty(t, options.WorldDrawers)

	drewCustom := false
	custom := DrawerFunc(func(*ebiten.Image, *application.Application) error {
		drewCustom = true
		return nil
	})
	WithWorldDrawers(custom)(options)
	sh = newShell(app, options, c)

	require.NoError(t, sh.worldDrawers.Draw(nil, app))
	require.True(t, drewCustom)
	require.Len(t, sh.worldDrawers, 1)
}
