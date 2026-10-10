package rendering

import (
	"fmt"
	"sort"
	"testing"

	"github.com/adm87/finch/pkg/ecs/components/renderer"
)

type keyParts struct {
	layer int16
	depth int16
	idx   int32
}

var benchmarkSink int

func lessByParts(a, b keyParts) bool {
	if a.layer != b.layer {
		return a.layer < b.layer
	}
	if a.depth != b.depth {
		return a.depth < b.depth
	}
	return a.idx < b.idx
}

func TestMakeSortedKeyOrder(t *testing.T) {
	items := []keyParts{
		{layer: 0, depth: 0, idx: 0},
		{layer: -1, depth: 0, idx: 0},
		{layer: 0, depth: -1, idx: 0},
		{layer: 0, depth: 0, idx: -1},
		{layer: 12, depth: -34, idx: 5},
		{layer: 12, depth: -34, idx: -5},
		{layer: 12, depth: 99, idx: -100},
		{layer: 32767, depth: 32767, idx: 2147483647},
		{layer: -32768, depth: -32768, idx: -2147483648},
	}

	expected := append([]keyParts(nil), items...)
	sort.Slice(expected, func(i, j int) bool {
		return lessByParts(expected[i], expected[j])
	})

	got := append([]keyParts(nil), items...)
	sort.Slice(got, func(i, j int) bool {
		ki := MakeSortKey(got[i].layer, got[i].depth, got[i].idx)
		kj := MakeSortKey(got[j].layer, got[j].depth, got[j].idx)
		return ki < kj
	})

	for i := range expected {
		if expected[i] != got[i] {
			t.Fatalf("sorted order mismatch at %d: expected %+v, got %+v", i, expected[i], got[i])
		}
	}
}

func TestRenderingQueueSortedOrderSynchronizesKeysAndCandidates(t *testing.T) {
	rq := NewRenderingQueue()

	type queued struct {
		layer int16
		depth int16
	}

	queuedCandidates := []queued{
		{layer: 3, depth: 40},
		{layer: -2, depth: 8},
		{layer: 3, depth: -10},
		{layer: 0, depth: 0},
		{layer: -2, depth: -3},
		{layer: 3, depth: -10}, // same layer/depth as index 2 to verify idx tie-break
		{layer: 3, depth: 40},  // same layer/depth as index 0 to verify idx tie-break
	}

	renderers := make([]*renderer.Renderer, 0, len(queuedCandidates))
	for i, c := range queuedCandidates {
		r := &renderer.Renderer{}
		r.SetID(uint64(i + 1))
		r.SetLayer(c.layer)
		r.SetDepth(c.depth)
		renderers = append(renderers, r)
		rq.Enqueue(nil, r)
	}

	wantOrder := make([]int, len(queuedCandidates))
	for i := range wantOrder {
		wantOrder[i] = i
	}

	sort.Slice(wantOrder, func(i, j int) bool {
		left := queuedCandidates[wantOrder[i]]
		right := queuedCandidates[wantOrder[j]]
		if left.layer != right.layer {
			return left.layer < right.layer
		}
		if left.depth != right.depth {
			return left.depth < right.depth
		}
		return wantOrder[i] < wantOrder[j]
	})

	keys := rq.SortedKeys()
	if len(keys) != len(wantOrder) {
		t.Fatalf("expected %d sorted keys, got %d", len(wantOrder), len(keys))
	}

	for sortedPos, originalIdx := range wantOrder {
		key := keys[sortedPos]
		if key.Layer() != queuedCandidates[originalIdx].layer {
			t.Fatalf("sorted position %d has layer %d, want %d", sortedPos, key.Layer(), queuedCandidates[originalIdx].layer)
		}
		if key.Depth() != queuedCandidates[originalIdx].depth {
			t.Fatalf("sorted position %d has depth %d, want %d", sortedPos, key.Depth(), queuedCandidates[originalIdx].depth)
		}
		if key.Index() != originalIdx {
			t.Fatalf("sorted position %d has index %d, want %d", sortedPos, key.Index(), originalIdx)
		}

		candidate := rq.CandidateAt(key.Index())
		if candidate == nil {
			t.Fatalf("candidate at index %d is nil", key.Index())
		}
		if candidate.Renderer != renderers[originalIdx] {
			t.Fatalf("candidate at index %d does not match enqueued renderer", key.Index())
		}
		if candidate.Renderer.Layer() != queuedCandidates[originalIdx].layer {
			t.Fatalf("candidate layer mismatch at index %d", key.Index())
		}
		if candidate.Renderer.Depth() != queuedCandidates[originalIdx].depth {
			t.Fatalf("candidate depth mismatch at index %d", key.Index())
		}
	}
}

func benchmarkRenderers(n int) []*renderer.Renderer {
	renderers := make([]*renderer.Renderer, 0, n)
	for i := 0; i < n; i++ {
		r := &renderer.Renderer{}
		r.SetID(uint64(i + 1))
		r.SetLayer(int16((i*37)%31 - 15))
		r.SetDepth(int16((i*101)%2001 - 1000))
		renderers = append(renderers, r)
	}
	return renderers
}

func BenchmarkMakeSortKey(b *testing.B) {
	var sk SortKey
	for i := 0; b.Loop(); i++ {
		layer := int16((i*37)%31 - 15)
		depth := int16((i*101)%2001 - 1000)
		sk = MakeSortKey(layer, depth, int32(i))
	}
	benchmarkSink = int(sk)
}

func BenchmarkRenderingQueueSortAndTraverse(b *testing.B) {
	for _, size := range []int{128, 1024, 4096} {
		b.Run(fmt.Sprintf("n=%d", size), func(b *testing.B) {
			renderers := benchmarkRenderers(size)
			rq := NewRenderingQueue()

			b.ReportAllocs()
			b.ResetTimer()

			sum := 0
			for i := 0; b.Loop(); i++ {
				rq.Clear()
				for _, r := range renderers {
					rq.Enqueue(nil, r)
				}

				for _, key := range rq.SortedKeys() {
					candidate := rq.CandidateAt(key.Index())
					if candidate == nil || candidate.Renderer == nil {
						b.Fatalf("invalid candidate at index %d", key.Index())
					}
					sum += int(candidate.Renderer.Layer()) + int(candidate.Renderer.Depth())
				}
			}

			benchmarkSink = sum
		})
	}
}
