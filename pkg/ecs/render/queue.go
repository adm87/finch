package render

import "math"

// Sort key bits, high to low: layer (8), depth (16), kind (8), job index (32).
const (
	layerShift = 56
	depthShift = 40
	kindShift  = 32
	indexMask  = 1<<32 - 1
	maxKinds   = 1<<8 - 1
)

type queue struct {
	keys []uint64
	jobs []Job
}

func (q *queue) reset() {
	q.keys = q.keys[:0]
	q.jobs = q.jobs[:0]
}

// key biases the signed layer and depth so they sort as unsigned.
func key(j *Job, index int) uint64 {
	layer := uint64(min(max(j.Layer, math.MinInt8), math.MaxInt8) - math.MinInt8)
	depth := uint64(min(max(j.Depth, math.MinInt16), math.MaxInt16) - math.MinInt16)
	return layer<<layerShift | depth<<depthShift | uint64(j.kind)<<kindShift | uint64(index) //nolint:gosec // kind <= maxKinds; index fits in 32 bits
}
