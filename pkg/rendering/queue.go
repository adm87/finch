package rendering

import (
	"fmt"
	"iter"
	"slices"

	"github.com/adm87/finch/pkg/errs"
)

const (
	int16SignMask uint16 = 0x8000
	int32SignMask uint32 = 0x80000000

	queueCap = 256

	// MaxJobs is how many jobs fit in one frame, since a sort key has 16 bits
	// for the job's index.
	MaxJobs = 1 << 16
)

// SortKey packs a job's layer, depth and queue index, so sorting keys sorts
// jobs by layer, then depth, then the order they were queued in.
type SortKey uint64

func MakeSortKey(layer int16, depth int32, index uint16) SortKey {
	normalizedLayer := uint64(uint16(layer) ^ int16SignMask)
	normalizedDepth := uint64(uint32(depth) ^ int32SignMask)
	return SortKey(normalizedLayer<<48 | normalizedDepth<<16 | uint64(index))
}

func (sk SortKey) Index() int {
	return int(uint16(sk))
}

func (sk SortKey) Depth() int32 {
	return int32(uint32(sk>>16) ^ int32SignMask)
}

func (sk SortKey) Layer() int16 {
	return int16(uint16(sk>>48) ^ int16SignMask)
}

// Queue holds a frame's jobs until they're drawn in sorted order.
type Queue struct {
	jobs []Job
	keys []SortKey
}

func NewQueue() *Queue {
	return &Queue{
		jobs: make([]Job, 0, queueCap),
		keys: make([]SortKey, 0, queueCap),
	}
}

func (rq *Queue) Enqueue(job Job) error {
	if len(rq.jobs) >= MaxJobs {
		return errs.InvalidOperation(fmt.Sprintf("render queue is full at %d jobs", MaxJobs))
	}
	rq.keys = append(rq.keys, MakeSortKey(job.Layer, job.Depth, uint16(len(rq.jobs))))
	rq.jobs = append(rq.jobs, job)
	return nil
}

func (rq *Queue) Len() int {
	return len(rq.jobs)
}

// Sorted yields the jobs by layer, then depth, then queue order.
func (rq *Queue) Sorted() iter.Seq[*Job] {
	return func(yield func(*Job) bool) {
		slices.Sort(rq.keys)
		for _, k := range rq.keys {
			if !yield(&rq.jobs[k.Index()]) {
				return
			}
		}
	}
}

func (rq *Queue) Clear() {
	clear(rq.jobs) // drop entry pointers so the queue doesn't keep them alive
	rq.jobs = rq.jobs[:0]
	rq.keys = rq.keys[:0]
}
