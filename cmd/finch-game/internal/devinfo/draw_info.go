package devinfo

import (
	"fmt"
	"runtime/metrics"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"

	"github.com/adm87/finch/pkg/engine"
)

type drawInfo struct {
	start      time.Time // start of the current second
	text       string    // rebuilt once a second so drawing it doesn't allocate
	samples    []metrics.Sample
	total      time.Duration // draw time so far this second
	frames     int
	lastAllocs uint64 // counts at the start of this second
	lastGCs    uint64
}

// DrawInfo prints FPS, CPU draw time and memory use, refreshed every second.
func DrawInfo() engine.DrawPass {
	return func(inner engine.Drawer) engine.Drawer {
		info := &drawInfo{
			start: time.Now(),
			samples: []metrics.Sample{
				{Name: "/memory/classes/heap/objects:bytes"},
				{Name: "/gc/heap/allocs:bytes"},
				{Name: "/gc/cycles/total:gc-cycles"},
			},
		}

		metrics.Read(info.samples)
		info.lastAllocs = info.samples[1].Value.Uint64()
		info.lastGCs = info.samples[2].Value.Uint64()

		return engine.DrawFunc(func(target *ebiten.Image) {
			begin := time.Now()
			inner.Draw(target)

			info.total += time.Since(begin)
			info.frames++

			if window := begin.Sub(info.start); window >= time.Second {
				avg := info.total / time.Duration(info.frames)
				budget := 100 * float64(info.total) / float64(window) // share of real time spent drawing

				// runtime/metrics reads without stopping the world, unlike ReadMemStats.
				metrics.Read(info.samples)
				heap := info.samples[0].Value.Uint64()
				allocs := info.samples[1].Value.Uint64()
				gcs := info.samples[2].Value.Uint64()
				secs := window.Seconds()

				info.text = fmt.Sprintf(
					"FPS: %.0f\nDraw (CPU): %s (%.1f%%)\nHeap: %.1f MB\nAlloc: %.1f KB/s\nGC: %.1f/s",
					ebiten.ActualFPS(),
					avg.Round(time.Microsecond),
					budget,
					float64(heap)/(1<<20),
					float64(allocs-info.lastAllocs)/(1<<10)/secs,
					float64(gcs-info.lastGCs)/secs,
				)

				info.start, info.total, info.frames = begin, 0, 0
				info.lastAllocs, info.lastGCs = allocs, gcs
			}

			ebitenutil.DebugPrint(target, info.text)
		})
	}
}
