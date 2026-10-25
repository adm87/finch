package engine

import (
	"errors"
	"testing"
	"time"

	"github.com/go-openapi/testify/require"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/yohamta/donburi"

	"github.com/adm87/finch/pkg/application"
	"github.com/adm87/finch/pkg/errs"
	"github.com/adm87/finch/pkg/geom"
	"github.com/adm87/finch/pkg/hexcolor"
	"github.com/adm87/finch/pkg/logger"
	"github.com/adm87/finch/pkg/transform"
	"github.com/adm87/finch/pkg/world"
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
	sh := newShell(app, *DefaultConfig(), options, c)
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
	require.ErrorIs(t, newShell(app, *DefaultConfig(), options, c).Update(), stop)
	require.False(t, late)
}

func TestDrawersDrawInOrderAndReportEveryError(t *testing.T) {
	var calls []string
	errA, errB := errors.New("a failed"), errors.New("b failed")
	drawer := func(name string, err error) Drawer {
		return DrawerFunc(func(*ebiten.Image, *DrawContext) error {
			calls = append(calls, name)
			return err
		})
	}

	err := Drawers{drawer("a", errA), drawer("b", nil), drawer("c", errB)}.Draw(nil, &DrawContext{})

	require.Equal(t, []string{"a", "b", "c"}, calls)
	require.ErrorIs(t, err, errA)
	require.ErrorIs(t, err, errB)
	require.NoError(t, Drawers(nil).Draw(nil, &DrawContext{}))
}

func TestShellDefaultsWorldDrawersWithoutChangingOptions(t *testing.T) {
	app, err := application.New(nil, logger.N)
	require.NoError(t, err)

	options := DefaultRunOptions()
	c, _ := newFakeClock(60)
	sh := newShell(app, *DefaultConfig(), options, c)

	require.Len(t, sh.worldDrawers, 1)
	require.Empty(t, options.WorldDrawers)

	drewCustom := false
	custom := DrawerFunc(func(*ebiten.Image, *DrawContext) error {
		drewCustom = true
		return nil
	})
	WithWorldDrawers(custom)(options)
	sh = newShell(app, *DefaultConfig(), options, c)

	require.NoError(t, sh.worldDrawers.Draw(nil, &DrawContext{}))
	require.True(t, drewCustom)
	require.Len(t, sh.worldDrawers, 1)
}

func TestShellSyncsWorldBeforeLateUpdate(t *testing.T) {
	app, err := application.New(nil, logger.N)
	require.NoError(t, err)

	w := app.World()
	entry := w.Create()
	transform.BoundsComponent.Get(entry).SetRect(geom.Rect{Width: 10, Height: 10})

	options := DefaultRunOptions()
	WithUpdate(func(*application.Application, float64) error {
		transform.TransformComponent.Get(entry).SetPosition(300, 300)
		return nil
	})(options)

	var seen []*donburi.Entry
	WithLateUpdate(func(app *application.Application, _ float64) error {
		seen = app.World().Query(geom.Rect{X: 300, Y: 300, Width: 10, Height: 10}, seen[:0])
		return nil
	})(options)

	c, _ := newFakeClock(60)
	require.NoError(t, newShell(app, *DefaultConfig(), options, c).Update())
	require.Len(t, seen, 1)
	require.Equal(t, entry.Entity(), seen[0].Entity())
}

func TestShellSyncsWorldAfterEveryPhase(t *testing.T) {
	app, err := application.New(nil, logger.N)
	require.NoError(t, err)

	entry := app.World().Create()
	transform.BoundsComponent.Get(entry).SetRect(geom.Rect{Width: 10, Height: 10})

	at := func(app *application.Application, x, y float64) bool {
		hits := app.World().Query(geom.Rect{X: x, Y: y, Width: 10, Height: 10}, nil)
		return len(hits) == 1 && hits[0].Entity() == entry.Entity()
	}
	moveTo := func(x, y float64) {
		transform.TransformComponent.Get(entry).SetPosition(x, y)
	}

	var seen []bool
	step := 0
	options := DefaultRunOptions()
	WithUpdate(func(*application.Application, float64) error {
		moveTo(100, 100)
		return nil
	})(options)
	WithFixedUpdate(func(app *application.Application, _ float64) error {
		step++
		if step == 1 {
			seen = append(seen, at(app, 100, 100))
			moveTo(200, 200)
			return nil
		}
		seen = append(seen, at(app, 200, 200))
		moveTo(300, 300)
		return nil
	})(options)
	WithLateUpdate(func(app *application.Application, _ float64) error {
		seen = append(seen, at(app, 300, 300))
		return nil
	})(options)

	c, ft := newFakeClock(10)
	sh := newShell(app, *DefaultConfig(), options, c)
	c.tick()
	ft.advance(200 * time.Millisecond)

	require.NoError(t, sh.Update())
	require.Equal(t, 2, step)
	require.Equal(t, []bool{true, true, true}, seen)
}

func TestShellGivesOverlaysTheWindowMatrix(t *testing.T) {
	cfg := *application.DefaultConfig()
	cfg.Screen.Width, cfg.Screen.Height = 320, 180
	app, err := application.New(&cfg, logger.N)
	require.NoError(t, err)

	app.Screen().Layout(400, 180)
	safeArea := app.Screen().SafeArea()
	require.NotZero(t, safeArea.Y)

	camera := app.World().Create(world.DontPartition)
	transform.TransformComponent.Get(camera).SetPosition(50, -20)
	app.View().SetCamera(camera)

	var worldMatrix, overlayMatrix geom.Matrix
	options := DefaultRunOptions()
	WithWorldDrawers(DrawerFunc(func(_ *ebiten.Image, ctx *DrawContext) error {
		worldMatrix = ctx.Matrix
		return nil
	}))(options)
	WithOverlayDrawers(DrawerFunc(func(_ *ebiten.Image, ctx *DrawContext) error {
		overlayMatrix = ctx.Matrix
		return nil
	}))(options)

	c, _ := newFakeClock(60)
	newShell(app, *DefaultConfig(), options, c).Draw(ebiten.NewImage(400, 180))

	bx, by := worldMatrix.Apply(10, 20)
	wx, wy := overlayMatrix.Apply(10, 20)
	require.InDelta(t, bx-safeArea.X, wx, 1e-9)
	require.InDelta(t, by-safeArea.Y, wy, 1e-9)
}

func TestShellWithoutCameraStillRunsOverlays(t *testing.T) {
	app, err := application.New(nil, logger.N)
	require.NoError(t, err)
	require.False(t, app.View().HasCamera())

	entry := app.World().Create()
	transform.BoundsComponent.Get(entry).SetRect(geom.Rect{Width: 10, Height: 10})
	require.NoError(t, app.World().Sync())

	drewWorld, drewOverlay := false, false
	var overlayEntries []*donburi.Entry
	var overlayRect geom.Rect

	options := DefaultRunOptions()
	WithWorldDrawers(DrawerFunc(func(*ebiten.Image, *DrawContext) error {
		drewWorld = true
		return nil
	}))(options)
	WithOverlayDrawers(DrawerFunc(func(_ *ebiten.Image, ctx *DrawContext) error {
		drewOverlay = true
		overlayEntries, overlayRect = ctx.Entries, ctx.View
		return nil
	}))(options)

	c, _ := newFakeClock(60)
	newShell(app, *DefaultConfig(), options, c).Draw(ebiten.NewImage(800, 600))

	require.False(t, drewWorld)
	require.True(t, drewOverlay)
	require.Empty(t, overlayEntries)
	require.Zero(t, overlayRect.Area())
}

func TestStartRunsStartupInOrder(t *testing.T) {
	var calls []string
	record := func(name string, err error) Startup {
		return func(*application.Application) error {
			calls = append(calls, name)
			return err
		}
	}

	options := DefaultRunOptions()
	WithStartup(record("a", nil), record("b", nil))(options)
	require.NoError(t, start(nil, options))
	require.Equal(t, []string{"a", "b"}, calls)

	calls = nil
	broken := errors.New("broken")
	options = DefaultRunOptions()
	WithStartup(record("a", broken), record("b", nil))(options)

	err := start(nil, options)
	require.ErrorIs(t, err, broken)
	require.ErrorContains(t, err, "startup:")
	require.Equal(t, []string{"a"}, calls)
}

func TestConfigValidate(t *testing.T) {
	require.NoError(t, DefaultConfig().Validate())

	cfg := *DefaultConfig()
	cfg.TickRate = 0
	cfg.Window.Width = 0
	cfg.Window.Height = -1
	err := cfg.Validate()
	require.ErrorIs(t, err, errs.ErrArgumentOutOfRange)
	require.ErrorContains(t, err, "tick_rate")
	require.ErrorContains(t, err, "window.width")
	require.ErrorContains(t, err, "window.height")
}

func TestShellClearsWithTheEngineColor(t *testing.T) {
	app, err := application.New(nil, logger.N)
	require.NoError(t, err)

	cfg := *DefaultConfig()
	cfg.ClearColor = hexcolor.Color{R: 10, G: 20, B: 30, A: 255}
	c, _ := newFakeClock(60)
	sh := newShell(app, cfg, DefaultRunOptions(), c)

	require.Equal(t, hexcolor.ToRGBA(cfg.ClearColor), sh.clearColor)
}
