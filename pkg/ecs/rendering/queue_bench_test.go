package rendering

import (
	"fmt"
	"slices"
	"testing"
)

// BenchmarkQueueSort pushes and sorts a frame's commands without drawing them.
func BenchmarkQueueSort(b *testing.B) {
	orders := []struct {
		name  string
		order func(i int) (layer, depth int16)
	}{
		{"uniform", func(int) (int16, int16) { return 0, 0 }},
		{"sorted", func(i int) (int16, int16) { return 0, int16(i) }},
		{"mixed", func(i int) (int16, int16) { return int16(i % 4), int16((i * 7919) % 1000) }},
	}
	for _, n := range []int{100, 1_000, 10_000} {
		for _, o := range orders {
			b.Run(fmt.Sprintf("commands=%d/%s", n, o.name), func(b *testing.B) {
				cmds := make([]Command, n)
				for i := range cmds {
					cmds[i].Layer, cmds[i].Depth = o.order(i)
				}

				var q Queue
				b.ReportAllocs()
				for b.Loop() {
					for _, c := range cmds {
						q.Push(c)
					}
					slices.Sort(q.keys)
					q.Reset()
				}
			})
		}
	}
}
