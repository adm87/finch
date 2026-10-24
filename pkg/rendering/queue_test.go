package rendering_test

import (
	"math"
	"slices"
	"testing"

	"github.com/go-openapi/testify/require"

	"github.com/adm87/finch/pkg/rendering"
)

func TestSortKeyRoundTrip(t *testing.T) {
	cases := []struct {
		layer, depth int16
		index        int32
	}{
		{0, 0, 0},
		{1, 2, 3},
		{-1, -1, -1},
		{math.MinInt16, math.MaxInt16, 0},
		{math.MaxInt16, math.MinInt16, math.MaxInt32},
		{-5, 7, math.MinInt32},
	}

	for _, tc := range cases {
		k := rendering.MakeSortKey(tc.layer, tc.depth, tc.index)
		require.Equal(t, tc.layer, k.Layer())
		require.Equal(t, tc.depth, k.Depth())
		require.Equal(t, int(tc.index), k.Index())
	}
}

func TestSortKeysOrderByLayerThenDepthThenIndex(t *testing.T) {
	want := []rendering.SortKey{
		rendering.MakeSortKey(math.MinInt16, math.MaxInt16, math.MaxInt32),
		rendering.MakeSortKey(-1, 0, 0),
		rendering.MakeSortKey(0, -1, 5),
		rendering.MakeSortKey(0, 0, -1),
		rendering.MakeSortKey(0, 0, 0),
		rendering.MakeSortKey(0, 0, 1),
		rendering.MakeSortKey(0, 1, 0),
		rendering.MakeSortKey(1, math.MinInt16, 0),
		rendering.MakeSortKey(math.MaxInt16, 0, 0),
	}

	got := slices.Clone(want)
	slices.Reverse(got)
	slices.Sort(got)
	require.Equal(t, want, got)
}

func newCandidate(layer, depth int16) rendering.Candidate {
	r := &rendering.Renderer{}
	r.SetLayer(layer)
	r.SetDepth(depth)
	return rendering.Candidate{Renderer: r}
}

func TestQueueDrawsInSortedOrder(t *testing.T) {
	q := rendering.NewQueue()
	q.Enqueue(newCandidate(1, 0))
	q.Enqueue(newCandidate(0, 5))
	q.Enqueue(newCandidate(0, -2))
	q.Enqueue(newCandidate(0, 5))

	var got [][2]int16
	var indexes []int
	for _, k := range q.SortedKeys() {
		rc := q.CandidateAt(k.Index())
		require.NotNil(t, rc)
		got = append(got, [2]int16{rc.Renderer.Layer(), rc.Renderer.Depth()})
		indexes = append(indexes, k.Index())
	}

	require.Equal(t, [][2]int16{{0, -2}, {0, 5}, {0, 5}, {1, 0}}, got)
	require.Equal(t, []int{2, 1, 3, 0}, indexes)
}

func TestQueueClear(t *testing.T) {
	q := rendering.NewQueue()
	q.Enqueue(newCandidate(0, 0))
	q.Clear()

	require.Empty(t, q.SortedKeys())
	require.Nil(t, q.CandidateAt(0))
}

func TestCandidateAtOutOfRange(t *testing.T) {
	q := rendering.NewQueue()
	q.Enqueue(newCandidate(0, 0))

	require.Nil(t, q.CandidateAt(-1))
	require.Nil(t, q.CandidateAt(1))
}
