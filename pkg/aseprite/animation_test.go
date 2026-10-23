package aseprite

import (
	"testing"

	"github.com/go-openapi/testify/require"

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

func newTestLibrary(direction Direction, frames int, duration int) LibraryInfo {
	atlas := &Atlas{
		Frames: make(Frames, frames),
		Meta: Meta{
			FrameTags: FrameTags{{Name: "Run", Direction: direction, From: 0, To: frames - 1}},
		},
	}
	for i := range atlas.Frames {
		atlas.Frames[i].Duration = duration
	}
	return LibraryInfo{Atlas: atlas}
}

func TestAdvanceAnimation(t *testing.T) {
	cases := []struct {
		name        string
		direction   Direction
		dt          float64
		wantFrame   int
		wantElapsed float64
	}{
		{"holds frame until its duration passes", DirectionForward, 0.05, 0, 50},
		{"steps once per duration", DirectionForward, 0.1, 1, 0},
		{"steps several frames in one update", DirectionForward, 0.25, 2, 50},
		{"wraps past the last frame", DirectionForward, 0.4, 1, 0},
		{"ping pong turns at the last frame", DirectionPingPong, 0.3, 1, 0},
		{"ignores non-positive dt", DirectionForward, 0, 0, 0},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			lib := newTestLibrary(tc.direction, 3, 100)
			anim := &Animator{animation: "Run"}
			img := &images.Image{}

			require.NoError(t, advanceAnimation(tc.dt, lib, anim, img))
			require.Equal(t, tc.wantFrame, img.Frame())
			require.InDelta(t, tc.wantElapsed, anim.Elapsed(), 1e-9)
		})
	}
}

func TestAdvanceAnimationUnknownTag(t *testing.T) {
	lib := newTestLibrary(DirectionForward, 3, 100)
	anim := &Animator{animation: "Missing"}
	img := &images.Image{}

	require.NoError(t, advanceAnimation(1, lib, anim, img))
	require.Equal(t, 0, img.Frame())
	require.Zero(t, anim.Elapsed())
}
