package rendering

import (
	"slices"

	"github.com/adm87/finch/pkg/ecs/components/renderer"
	"github.com/yohamta/donburi"
)

const (
	int16SignMask uint16 = 0x8000
	int32SignMask uint32 = 0x80000000

	queueCap = 256
)

type SortKey uint64

func MakeSortKey(layer, depth int16, i int32) SortKey {
	normalizedLayer := uint64(uint16(layer) ^ int16SignMask)
	normalizedDepth := uint64(uint16(depth) ^ int16SignMask)
	normalizedIdx := uint64(uint32(i) ^ int32SignMask)
	return SortKey(normalizedLayer<<48 | normalizedDepth<<32 | normalizedIdx)
}

func (sk SortKey) Index() int {
	return int(int32(uint32(sk) ^ int32SignMask))
}

func (sk SortKey) Depth() int16 {
	return int16(uint16(sk>>32) ^ int16SignMask)
}

func (sk SortKey) Layer() int16 {
	return int16(uint16(sk>>48) ^ int16SignMask)
}

type RenderingQueue struct {
	candidates []RenderCandidate
	sortedKeys []SortKey
}

func NewRenderingQueue() *RenderingQueue {
	return &RenderingQueue{
		candidates: make([]RenderCandidate, 0, queueCap),
		sortedKeys: make([]SortKey, 0, queueCap),
	}
}

func (rq *RenderingQueue) Enqueue(entry *donburi.Entry, renderer *renderer.Renderer) {
	rq.candidates = append(rq.candidates, RenderCandidate{
		Entry:    entry,
		Renderer: renderer,
	})
	rq.sortedKeys = append(rq.sortedKeys, MakeSortKey(
		renderer.Layer(),
		renderer.Depth(),
		int32(len(rq.candidates)-1),
	))
}

func (rq *RenderingQueue) CandidateAt(i int) *RenderCandidate {
	if i < 0 || i >= len(rq.candidates) {
		return nil
	}
	return &rq.candidates[i]
}

func (rq *RenderingQueue) SortedKeys() []SortKey {
	slices.Sort(rq.sortedKeys)
	return rq.sortedKeys
}

func (rq *RenderingQueue) Clear() {
	rq.candidates = rq.candidates[:0]
	rq.sortedKeys = rq.sortedKeys[:0]
}
