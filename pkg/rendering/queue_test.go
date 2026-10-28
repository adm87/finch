package rendering_test

import (
	"math"
	"slices"
	"testing"

	"github.com/go-openapi/testify/require"

	"github.com/adm87/finch/pkg/errs"
	"github.com/adm87/finch/pkg/rendering"
)

func TestSortKeyRoundTrip(t *testing.T) {
	cases := []struct {
		depth int32
		layer int16
		index uint16
	}{
		{0, 0, 0},
		{2, 1, 3},
		{-1, -1, math.MaxUint16},
		{math.MaxInt32, math.MinInt16, 0},
		{math.MinInt32, math.MaxInt16, math.MaxUint16},
		{7, -5, 12},
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
		rendering.MakeSortKey(math.MinInt16, math.MaxInt32, math.MaxUint16),
		rendering.MakeSortKey(-1, 0, 0),
		rendering.MakeSortKey(0, math.MinInt32, 5),
		rendering.MakeSortKey(0, -1, 5),
		rendering.MakeSortKey(0, 0, 0),
		rendering.MakeSortKey(0, 0, 1),
		rendering.MakeSortKey(0, 1, 0),
		rendering.MakeSortKey(0, math.MaxInt32, 0),
		rendering.MakeSortKey(1, math.MinInt32, 0),
		rendering.MakeSortKey(math.MaxInt16, 0, 0),
	}

	got := slices.Clone(want)
	slices.Reverse(got)
	slices.Sort(got)
	require.Equal(t, want, got)
}

func TestQueueYieldsInSortedOrder(t *testing.T) {
	q := rendering.NewQueue()
	for i, job := range []rendering.Job{
		{Layer: 1, Depth: 0},
		{Layer: 0, Depth: 5},
		{Layer: 0, Depth: -2},
		{Layer: 0, Depth: 5},
	} {
		job.Pass = i
		require.NoError(t, q.Enqueue(job))
	}

	var passes []int
	for job := range q.Sorted() {
		passes = append(passes, job.Pass)
	}
	require.Equal(t, []int{2, 1, 3, 0}, passes)
}

func TestQueueClear(t *testing.T) {
	q := rendering.NewQueue()
	require.NoError(t, q.Enqueue(rendering.Job{}))
	q.Clear()

	require.Zero(t, q.Len())
	for range q.Sorted() {
		t.Fatal("cleared queue yielded a job")
	}
}

func TestQueueRejectsMoreThanMaxJobs(t *testing.T) {
	q := rendering.NewQueue()
	for range rendering.MaxJobs {
		require.NoError(t, q.Enqueue(rendering.Job{}))
	}
	require.ErrorIs(t, q.Enqueue(rendering.Job{}), errs.ErrInvalidOperation)
}
