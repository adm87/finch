package rendering

import (
	"slices"
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

type Queue struct {
	candidates []Candidate
	sortedKeys []SortKey
}

func NewQueue() *Queue {
	return &Queue{
		candidates: make([]Candidate, 0, queueCap),
		sortedKeys: make([]SortKey, 0, queueCap),
	}
}

func (rq *Queue) Enqueue(rc Candidate) {
	rq.candidates = append(rq.candidates, rc)
	rq.sortedKeys = append(rq.sortedKeys, MakeSortKey(
		rc.Renderer.Layer(),
		rc.Renderer.Depth(),
		int32(len(rq.candidates)-1),
	))
}

func (rq *Queue) CandidateAt(i int) *Candidate {
	if i < 0 || i >= len(rq.candidates) {
		return nil
	}
	return &rq.candidates[i]
}

func (rq *Queue) SortedKeys() []SortKey {
	slices.Sort(rq.sortedKeys)
	return rq.sortedKeys
}

func (rq *Queue) Clear() {
	rq.candidates = rq.candidates[:0]
	rq.sortedKeys = rq.sortedKeys[:0]
}
