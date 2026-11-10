package hashgrid_test

import (
	"fmt"
	"math/rand/v2"
	"testing"

	"github.com/adm87/finch/pkg/geom"
	"github.com/adm87/finch/pkg/hashgrid"
)

// The world spans -2048..2048, so every axis crosses zero.
const (
	world    = 4096.0
	origin   = -world / 2
	cellSize = 64
)

type body struct {
	rect   geom.Rect
	vx, vy float64
}

func bodies(n int) []body {
	r := rand.New(rand.NewPCG(1, 2))
	bs := make([]body, n)
	for i := range bs {
		s := 8 + r.Float64()*40
		bs[i] = body{
			rect: geom.Rect{X: origin + r.Float64()*(world-s), Y: origin + r.Float64()*(world-s), Width: s, Height: s},
			vx:   r.Float64()*8 - 4,
			vy:   r.Float64()*8 - 4,
		}
	}
	return bs
}

// step moves every body and bounces it off the world's edges.
func step(bs []body) {
	for i := range bs {
		b := &bs[i]
		b.rect.X += b.vx
		b.rect.Y += b.vy
		if b.rect.X < origin || b.rect.X+b.rect.Width > origin+world {
			b.vx = -b.vx
		}
		if b.rect.Y < origin || b.rect.Y+b.rect.Height > origin+world {
			b.vy = -b.vy
		}
	}
}

func queries(n int, w, h float64) []geom.Rect {
	r := rand.New(rand.NewPCG(3, 4))
	qs := make([]geom.Rect, n)
	for i := range qs {
		qs[i] = geom.Rect{X: origin + r.Float64()*(world-w), Y: origin + r.Float64()*(world-h), Width: w, Height: h}
	}
	return qs
}

func fill(bs []body) (*hashgrid.Grid[int], []hashgrid.Key) {
	g := hashgrid.New[int](cellSize, len(bs))
	keys := make([]hashgrid.Key, len(bs))
	for i, b := range bs {
		keys[i] = g.Insert(b.rect, i)
	}
	return g, keys
}

func BenchmarkInsert(b *testing.B) {
	for _, n := range []int{1_000, 10_000} {
		bs := bodies(n)
		b.Run(fmt.Sprintf("items=%d", n), func(b *testing.B) {
			b.ReportAllocs()
			for b.Loop() {
				fill(bs)
			}
		})
	}
}

// One frame: every body moves a few units and is updated.
func BenchmarkMoveAll(b *testing.B) {
	for _, n := range []int{1_000, 10_000} {
		b.Run(fmt.Sprintf("items=%d", n), func(b *testing.B) {
			bs := bodies(n)
			g, keys := fill(bs)
			b.ReportAllocs()
			for b.Loop() {
				step(bs)
				for i := range bs {
					g.Update(keys[i], bs[i].rect)
				}
			}
		})
	}
}

// A tenth of the bodies are deleted and reinserted elsewhere each frame.
func BenchmarkChurn(b *testing.B) {
	const n = 10_000
	bs := bodies(n)
	g, keys := fill(bs)
	i := 0
	b.ReportAllocs()
	for b.Loop() {
		for range n / 10 {
			g.Delete(keys[i])
			bs[i].rect.X = -bs[i].rect.X - bs[i].rect.Width
			keys[i] = g.Insert(bs[i].rect, i)
			i = (i + 1) % n
		}
	}
}

// A small area check, a 1280x720 view, and the whole world.
func BenchmarkQuery(b *testing.B) {
	g, _ := fill(bodies(10_000))
	sizes := []struct {
		name  string
		w, h  float64
		count int
	}{
		{"small", 128, 128, 1000},
		{"view", 1280, 720, 10},
		{"world", world, world, 1},
	}
	for _, s := range sizes {
		qs := queries(s.count, s.w, s.h)
		b.Run(s.name, func(b *testing.B) {
			var dst []int
			b.ReportAllocs()
			for b.Loop() {
				for _, q := range qs {
					dst = g.Query(q, dst[:0])
				}
			}
		})
	}
}

func BenchmarkRange(b *testing.B) {
	g, _ := fill(bodies(10_000))
	qs := queries(1000, 128, 128)
	var dst []int
	b.ReportAllocs()
	for b.Loop() {
		for _, q := range qs {
			dst = dst[:0]
			for _, v := range g.Range(q) {
				dst = append(dst, v)
			}
		}
	}
}
