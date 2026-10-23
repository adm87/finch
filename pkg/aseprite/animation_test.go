package aseprite

import (
	"testing"

	"github.com/go-openapi/testify/require"
	"github.com/yohamta/donburi"

	"github.com/adm87/finch/pkg/images"
)

func TestNextFrame(t *testing.T) {
	cases := []struct {
		name     string
		start    int
		reverse  bool
		pingPong bool
		want     []int
	}{
		{"forward wraps", 0, false, false, []int{1, 2, 0, 1}},
		{"reverse wraps", 2, true, false, []int{1, 0, 2, 1}},
		{"ping pong bounces", 0, false, true, []int{1, 2, 1, 0, 1}},
		{"ping pong reverse bounces", 2, true, true, []int{1, 0, 1, 2, 1}},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			idx, reversing := tc.start, tc.reverse
			var got []int
			for range tc.want {
				idx, reversing = nextFrame(idx, 3, tc.reverse, tc.pingPong, reversing)
				got = append(got, idx)
			}
			require.Equal(t, tc.want, got)
		})
	}
}

// newTestLibrary returns an atlas with a single "Run" tag starting at frame from.
func newTestLibrary(direction Direction, from int, durations ...int) LibraryInfo {
	atlas := &Atlas{
		Frames: make(Frames, from+len(durations)+1),
		Meta: Meta{
			FrameTags: FrameTags{{Name: "Run", Direction: direction, From: from, To: from + len(durations) - 1}},
		},
	}
	for i, d := range durations {
		atlas.Frames[from+i].Duration = d
	}
	return LibraryInfo{Atlas: atlas}
}

func TestAdvanceAnimation(t *testing.T) {
	cases := []struct {
		name        string
		direction   Direction
		durations   []int
		dt          float64
		wantFrame   int
		wantElapsed float64
	}{
		{"holds frame until its duration passes", DirectionForward, []int{100, 100, 100}, 0.05, 0, 50},
		{"steps once per duration", DirectionForward, []int{100, 100, 100}, 0.1, 1, 0},
		{"steps several frames in one update", DirectionForward, []int{100, 100, 100}, 0.25, 2, 50},
		{"wraps past the last frame", DirectionForward, []int{100, 100, 100}, 0.4, 1, 0},
		{"ping pong turns at the last frame", DirectionPingPong, []int{100, 100, 100}, 0.3, 1, 0},
		{"uses each frame's own duration", DirectionForward, []int{100, 300, 100}, 0.35, 1, 250},
		{"skips whole loops of mixed durations", DirectionForward, []int{100, 300, 100}, 1.35, 1, 250},
		{"reverse starts on the last frame", DirectionReverse, []int{100, 100, 100}, 0, 2, 0},
		{"reverse steps backwards", DirectionReverse, []int{100, 100, 100}, 0.1, 1, 0},
		{"ping pong reverse starts on the last frame", DirectionPingPongReverse, []int{100, 100, 100}, 0.1, 1, 0},
		{"single frame keeps its place", DirectionForward, []int{100}, 0.25, 0, 50},
		{"frame with no duration holds", DirectionForward, []int{100, 0, 100}, 5, 1, 4900},
		{"ignores non-positive dt", DirectionForward, []int{100, 100, 100}, 0, 0, 0},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			const from = 5
			lib := newTestLibrary(tc.direction, from, tc.durations...)
			anim := &Animator{animation: "Run", loops: LoopForever}
			img := &images.Image{}

			require.NoError(t, advanceAnimation(tc.dt, lib, anim, img))
			require.Equal(t, tc.wantFrame, anim.Frame())
			require.Equal(t, from+tc.wantFrame, img.Frame())
			require.InDelta(t, tc.wantElapsed, anim.Elapsed(), 1e-6)
		})
	}
}

func TestAdvanceAnimationRestarts(t *testing.T) {
	lib := newTestLibrary(DirectionForward, 0, 100, 100, 100)
	lib.Atlas.Meta.FrameTags = append(lib.Atlas.Meta.FrameTags,
		FrameTag{Name: "Back", Direction: DirectionReverse, From: 0, To: 2})
	anim := &Animator{loops: LoopForever}
	anim.SetAnimation("Run")
	img := &images.Image{}

	require.NoError(t, advanceAnimation(0.15, lib, anim, img))
	require.Equal(t, 1, anim.Frame())

	anim.SetAnimation("Run")
	require.NoError(t, advanceAnimation(0, lib, anim, img))
	require.Equal(t, 1, anim.Frame())
	require.InDelta(t, 50, anim.Elapsed(), 1e-6)

	anim.SetAnimation("Back")
	require.NoError(t, advanceAnimation(0, lib, anim, img))
	require.Equal(t, 2, anim.Frame())
	require.Zero(t, anim.Elapsed())

	require.NoError(t, advanceAnimation(0.05, lib, anim, img))
	anim.SetLibrary(anim.Library())
	require.NoError(t, advanceAnimation(0, lib, anim, img))
	require.Equal(t, 2, anim.Frame())
	require.Zero(t, anim.Elapsed())
}

func TestAdvanceAnimationUnknownTag(t *testing.T) {
	lib := newTestLibrary(DirectionForward, 0, 100, 100, 100)
	anim := &Animator{animation: "Missing", loops: LoopForever}
	img := &images.Image{}

	require.NoError(t, advanceAnimation(1, lib, anim, img))
	require.Equal(t, 0, img.Frame())
	require.Zero(t, anim.Elapsed())
}

func TestAdvanceAnimationLoops(t *testing.T) {
	cases := []struct {
		name         string
		direction    Direction
		durations    []int
		loops        int
		dt           float64
		wantFrame    int
		wantFinished bool
	}{
		{"once is still playing before the last frame ends", DirectionForward, []int{100, 100, 100}, 0, 0.25, 2, false},
		{"once holds the last frame", DirectionForward, []int{100, 100, 100}, 0, 0.3, 2, true},
		{"once holds through a long update", DirectionForward, []int{100, 100, 100}, 0, 10, 2, true},
		{"one loop plays a second cycle", DirectionForward, []int{100, 100, 100}, 1, 0.55, 2, false},
		{"one loop finishes after two cycles", DirectionForward, []int{100, 100, 100}, 1, 0.6, 2, true},
		{"several loops finish in one long update", DirectionForward, []int{100, 300, 100}, 3, 100, 2, true},
		{"reverse once holds the first frame", DirectionReverse, []int{100, 100, 100}, 0, 1, 0, true},
		{"ping pong once comes back before finishing", DirectionPingPong, []int{100, 100, 100}, 0, 0.35, 1, false},
		{"ping pong once holds where it started", DirectionPingPong, []int{100, 100, 100}, 0, 0.4, 0, true},
		{"ping pong reverse once holds where it started", DirectionPingPongReverse, []int{100, 100, 100}, 0, 0.4, 2, true},
		{"ping pong one loop plays out and back twice", DirectionPingPong, []int{100, 100, 100}, 1, 0.75, 1, false},
		{"ping pong loops finish in one long update", DirectionPingPong, []int{100, 100, 100}, 2, 100, 0, true},
		{"single frame once", DirectionForward, []int{100}, 0, 0.1, 0, true},
		{"single frame ping pong once", DirectionPingPong, []int{100}, 0, 0.1, 0, true},
		{"loop forever never finishes", DirectionForward, []int{100, 100, 100}, LoopForever, 100, 1, false},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			lib := newTestLibrary(tc.direction, 0, tc.durations...)
			anim := &Animator{animation: "Run"}
			anim.SetLoops(tc.loops)
			img := &images.Image{}

			require.NoError(t, advanceAnimation(tc.dt, lib, anim, img))
			require.Equal(t, tc.wantFrame, anim.Frame())
			require.Equal(t, tc.wantFinished, anim.Finished())
		})
	}
}

func TestAnimatorLoopChanges(t *testing.T) {
	lib := newTestLibrary(DirectionForward, 0, 100, 100, 100)
	img := &images.Image{}

	t.Run("more loops resume a finished animation", func(t *testing.T) {
		anim := &Animator{animation: "Run"}
		require.NoError(t, advanceAnimation(0.3, lib, anim, img))
		require.True(t, anim.Finished())

		anim.SetLoops(1)
		require.False(t, anim.Finished())
		require.NoError(t, advanceAnimation(0.05, lib, anim, img))
		require.Equal(t, 0, anim.Frame())
		require.NoError(t, advanceAnimation(0.25, lib, anim, img))
		require.True(t, anim.Finished())
		require.Equal(t, 2, anim.Frame())
	})

	t.Run("fewer loops finish at the end of the current cycle", func(t *testing.T) {
		anim := &Animator{animation: "Run"}
		anim.SetLoops(3)
		require.NoError(t, advanceAnimation(0.45, lib, anim, img))
		require.Equal(t, 1, anim.Frame())

		anim.SetLoops(0)
		require.False(t, anim.Finished())
		require.NoError(t, advanceAnimation(0.1, lib, anim, img))
		require.False(t, anim.Finished())
		require.NoError(t, advanceAnimation(0.1, lib, anim, img))
		require.True(t, anim.Finished())
		require.Equal(t, 2, anim.Frame())
	})

	t.Run("restart plays a finished animation again", func(t *testing.T) {
		anim := &Animator{animation: "Run"}
		require.NoError(t, advanceAnimation(1, lib, anim, img))
		require.True(t, anim.Finished())

		anim.Restart()
		require.NoError(t, advanceAnimation(0, lib, anim, img))
		require.False(t, anim.Finished())
		require.Equal(t, 0, anim.Frame())
	})

	t.Run("new animations default to looping forever", func(t *testing.T) {
		w := donburi.NewWorld()
		anim, ok := GetAnimator(w.Entry(w.Create(AnimatorComponent)))
		require.True(t, ok)
		require.Equal(t, LoopForever, anim.Loops())
	})

	t.Run("loop counts below forever clamp to forever", func(t *testing.T) {
		anim := &Animator{}
		anim.SetLoops(-5)
		require.Equal(t, LoopForever, anim.Loops())
	})
}
